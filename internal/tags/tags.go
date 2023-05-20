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

type TagWriter interface {
	WriteTag(*ixml.XmlWriter, string) error //Writes the tag's content to an xml file
	CheckTag() error                        //Returns an error if the tag is missing mandatory data
}

type Actor struct {
	Name      string
	Character string
}

func (r *Actor) CheckTag() error {
	if len(r.Name) < 1 {
		return fmt.Errorf("Name not set for actor")
	}
	return nil
}

func (r *Actor) WriteTag(xw *ixml.XmlWriter, name string) error {
	if err := r.CheckTag(); err != nil {
		return err
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

func (r UniLingual) CheckTag() error {
	if len(r) < 1 {
		return fmt.Errorf("Tag is empty")
	}
	return nil
}

func (r UniLingual) WriteTag(xw *ixml.XmlWriter, name string) error {
	if err := r.CheckTag(); err != nil {
		return err
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

func (r *MultiLingual) CheckTag() error {
	if len(r.Text) < 1 {
		return fmt.Errorf("Tag is empty")
	}
	return nil
}

func (r *MultiLingual) WriteTag(xw *ixml.XmlWriter, name string) error {
	if err := r.CheckTag(); err != nil {
		return err
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

func (r *Country) CheckTag() error {
	if len(r.Name) < 1 {
		return fmt.Errorf("Missing country name")
	}
	if r.IsEmpty() {
		return fmt.Errorf("Country does not contain any payload")
	}
	return nil
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

func (r *Movie) CheckTag() error {
	return nil
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

// Writes the matroska tags as an xml file
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
			if err := tw.CheckTag(); err != nil {
				global.Log.Debug(fmt.Sprintf("writeTaggedFields: Did not write tag \"%s\": %s", tag, err))
				return nil
			}
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
		if slice.Len() < 1 {
			global.Log.Debug(fmt.Sprintf("writeTaggedFields: Skipping empty slice tag \"%s\"", tag))
			return nil
		}
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
