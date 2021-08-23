//This file is part of imdb2mkvtags ©2021 Jörg Walter

package tags

import (
	"encoding/xml"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/global"
	"github.com/jwdev42/imdb2mkvtags/internal/util/dynamic"
	ixml "github.com/jwdev42/imdb2mkvtags/internal/xml"
	"io"
	"reflect"
)

type EmptyTag string

func (r EmptyTag) Error() string {
	return string(r)
}

func NewEmptyTag(tag string) EmptyTag {
	return EmptyTag(fmt.Sprintf("Tag is empty or incomplete: %s", tag))
}

type TagWriter interface {
	WriteTag(*ixml.XmlWriter, string) error
}

type Actor struct {
	Name      string
	Character string
}

func (r *Actor) WriteTag(xw *ixml.XmlWriter, name string) error {
	if len(r.Name) < 1 {
		return NewEmptyTag(name)
	}
	var fc func(*ixml.XmlWriter) error
	if r.Character != "" {
		character := make([][2]string, 0, 2)
		character = append(character, [2]string{"Name", "CHARACTER"})
		character = append(character, [2]string{"String", r.Character})
		fc = ixml.GenWriteTagWithSubtags("Simple", character, nil)
	}
	actor := make([][2]string, 0, 2)
	actor = append(actor, [2]string{"Name", name})
	actor = append(actor, [2]string{"String", r.Name})
	return ixml.WriteTagWithSubtags(xw, "Simple", actor, fc)
}

type UniLingual string

func (r UniLingual) WriteTag(xw *ixml.XmlWriter, name string) error {
	if len(r) < 1 {
		return NewEmptyTag(name)
	}
	subtags := make([][2]string, 0, 2)
	subtags = append(subtags, [2]string{"Name", name})
	subtags = append(subtags, [2]string{"String", string(r)})
	return ixml.WriteTagWithSubtags(xw, "Simple", subtags, nil)
}

type MultiLingual struct {
	Text string
	Lang string
}

func (r *MultiLingual) WriteTag(xw *ixml.XmlWriter, name string) error {
	if len(r.Text) < 1 {
		return NewEmptyTag(name)
	}
	subtags := make([][2]string, 0, 3)
	subtags = append(subtags, [2]string{"Name", name})
	subtags = append(subtags, [2]string{"String", r.Text})
	if r.Lang != "" {
		subtags = append(subtags, [2]string{"TagLanguageIETF", r.Lang})
	}
	return ixml.WriteTagWithSubtags(xw, "Simple", subtags, nil)
}

type Country struct {
	Name      string
	nonempty  bool
	LawRating UniLingual `mkv:"LAW_RATING"`
}

func (r *Country) SetFieldCallback(name string, callback interface{}) {
	if err := dynamic.SetStructFieldCallback(name, r, callback); err != nil {
		global.Log.Error(fmt.Errorf("Country: Could not set field \"%s\": %s", name, err))
	} else {
		r.nonempty = true
	}
}

func (r *Country) IsEmpty() bool {
	return !r.nonempty
}

func (r *Country) WriteTag(xw *ixml.XmlWriter, name string) error {
	if name != "COUNTRY" {
		panic("Country's method \"WriteTag\" must be called with name == \"COUNTRY\"")
	}
	if err := xw.EncodeToken(ixml.NewStartElementSimple("Simple")); err != nil {
		return err
	}

	//Start Name tag
	if err := xw.EncodeToken(ixml.NewStartElementSimple("Name")); err != nil {
		return err
	}
	if err := xw.WriteText([]byte(name)); err != nil {
		return err
	}
	if err := xw.CloseElement(); err != nil {
		return err
	}
	//End Name tag

	//Start String tag
	if err := xw.EncodeToken(ixml.NewStartElementSimple("String")); err != nil {
		return err
	}
	if err := xw.WriteText([]byte(r.Name)); err != nil {
		return err
	}
	if err := xw.CloseElement(); err != nil {
		return err
	}
	//End String tag

	if err := writeTaggedFields(xw, r); err != nil {
		return err
	}

	if err := xw.CloseElement(); err != nil {
		return err
	}
	return nil
}

type Movie struct {
	Actors       []Actor        `mkv:"ACTOR"`
	Countries    []*Country     `mkv:"COUNTRY"`
	DateReleased UniLingual     `mkv:"DATE_RELEASED"`
	DateTagged   UniLingual     `mkv:"DATE_TAGGED"`
	Directors    []UniLingual   `mkv:"DIRECTOR"`
	Genres       []MultiLingual `mkv:"GENRE"`
	Imdb         UniLingual     `mkv:"IMDB"`
	Keywords     []MultiLingual `mkv:"KEYWORDS"`
	Producers    []UniLingual   `mkv:"PRODUCER"`
	Synopses     []MultiLingual `mkv:"SYNOPSIS"`
	Titles       []MultiLingual `mkv:"TITLE"`
	Writers      []UniLingual   `mkv:"WRITTEN_BY"`
}

func (r *Movie) SetFieldCallback(name string, callback interface{}) {
	if err := dynamic.SetStructFieldCallback(name, r, callback); err != nil {
		global.Log.Error(fmt.Errorf("Movie: Could not set field \"%s\": %s", name, err))
	}
}

func (r *Movie) WriteTag(xw *ixml.XmlWriter) error {

	//Write "Header" with TargetTypeValue=50
	{
		if err := xw.EncodeTokens(
			ixml.NewStartElementSimple("Tag"), ixml.NewStartElementSimple("Targets"),
			ixml.NewStartElementSimple("TargetTypeValue")); err != nil {
			return err
		}
		if err := xw.WriteText([]byte("50")); err != nil {
			return err
		}
		if err := xw.CloseElement(); err != nil {
			return err
		}
		if err := xw.CloseElement(); err != nil {
			return err
		}
	}

	if err := writeTaggedFields(xw, r); err != nil {
		return err
	}

	if err := xw.CloseElement(); err != nil {
		return err
	}

	return nil
}

//Writes the matroska tags as an xml file
func WriteTags(w io.Writer, tagWriter func(*ixml.XmlWriter) error) error {
	xw := ixml.NewXmlWriter(w)
	xw.Indent("", "\t")
	if err := xw.EncodeToken(xml.ProcInst{Target: "xml", Inst: []byte(`version="1.0" encoding="UTF-8"`)}); err != nil {
		return err
	}
	if err := xw.EncodeToken(ixml.NewStartElementSimple("Tags")); err != nil {
		return err
	}
	if err := tagWriter(xw); err != nil {
		return err
	}
	if err := xw.CloseAllElements(); err != nil {
		return err
	}
	if err := xw.Flush(); err != nil {
		return err
	}
	if _, err := w.Write([]byte{'\n'}); err != nil {
		return err
	}
	return nil
}

func writeTaggedFields(xw *ixml.XmlWriter, rec interface{}) error {

	write := func(field reflect.Value, tag string) error {
		i := field.Interface()
		if tw, ok := i.(TagWriter); ok {
			if err := tw.WriteTag(xw, tag); err != nil {
				return err
			}
		} else {
			panic("Field is not a TagWriter")
		}
		return nil
	}

	ptr := func(v reflect.Value) reflect.Value {
		if v.Type().Kind() == reflect.Ptr {
			return v
		}
		return v.Addr()
	}

	iterate := func(slice reflect.Value, tag string) error {
		for i := 0; i < slice.Len(); i++ {
			e := slice.Index(i)
			if err := write(ptr(e), tag); err != nil {
				return err
			}
		}
		return nil
	}

	fields, tags := dynamic.FieldsByStructTag("mkv", rec)
	for i, field := range fields {
		if field.Kind() == reflect.Slice {
			if field.Len() < 1 {
				global.Log.Debug(NewEmptyTag(tags[i]))
				continue
			}
			if err := iterate(field, tags[i]); err != nil {
				return err
			}
		} else {
			if err := write(ptr(field), tags[i]); err != nil {
				return err
			}
		}
	}
	return nil
}
