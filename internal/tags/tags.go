//This file is part of imdb2mkvtags ©2021 Jörg Walter

package tags

import (
	"encoding/xml"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/global"
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

type Movie struct {
	Actors        []Actor        `mkv:"ACTOR"`
	ContentRating []MultiLingual `mkv:"LAW_RATING"`
	Directors     []UniLingual   `mkv:"DIRECTOR"`
	Genres        []MultiLingual `mkv:"GENRE"`
	Imdb          UniLingual     `mkv:"IMDB"`
	Keywords      UniLingual     `mkv:"KEYWORDS"`
	Producers     []UniLingual   `mkv:"PRODUCER"`
	ReleaseDate   UniLingual     `mkv:"DATE_RELEASED"`
	Synopses      []MultiLingual `mkv:"SYNOPSIS"`
	Titles        []MultiLingual `mkv:"TITLE"`
	Writers       []UniLingual   `mkv:"WRITTEN_BY"`
}

func (r *Movie) SetField(name string, data interface{}) {
	val := reflect.Indirect(reflect.ValueOf(r))
	fieldVal := val.FieldByName(name)
	if fieldVal == (reflect.Value{}) {
		panic(fmt.Errorf(`Field "%s" not a member of struct Movie`, name))
	}
	fieldVal.Set(reflect.ValueOf(data))
}

func (r *Movie) writeFields(xw *ixml.XmlWriter) error {

	hasMethodWriteTag := func(v reflect.Value) bool {
		_, ok := v.Type().MethodByName("WriteTag")
		return ok
	}

	hasMkvTag := func(desc reflect.StructField) bool {
		if desc.Tag.Get("mkv") != "" {
			return true
		}
		return false
	}

	fieldWriter := func(v reflect.Value) reflect.Value {
		if hasMethodWriteTag(v) {
			return v
		} else if hasMethodWriteTag(v.Addr()) {
			return v.Addr()
		}
		return reflect.Value{}
	}

	writeIfEligible := func(v reflect.Value, tagName string) error {
		fw := fieldWriter(v)
		if fw == (reflect.Value{}) {
			panic("Object does not have a Method \"WriteTag\"")
		}
		if err := r.writeField(xw, fw, tagName); err != nil {
			if _, ok := err.(EmptyTag); ok {
				global.Log.Debug(err)
			} else {
				return err
			}
		}
		return nil
	}

	iterate := func(slice reflect.Value, tagName string) error {
		for i := 0; i < slice.Len(); i++ {
			e := slice.Index(i)
			if err := writeIfEligible(e, tagName); err != nil {
				return err
			}
		}
		return nil
	}

	inspectField := func(fieldVal reflect.Value, fieldDesc reflect.StructField) error {
		mkvTag := fieldDesc.Tag.Get("mkv")
		if mkvTag == "" {
			panic("function requires an non-empty mkv struct tag")
		}
		if fieldVal.Kind() == reflect.Slice {
			return iterate(fieldVal, mkvTag)
		}
		return writeIfEligible(fieldVal, mkvTag)
	}

	structVal := reflect.Indirect(reflect.ValueOf(r))
	for i := 0; i < structVal.NumField(); i++ {
		fieldDesc := structVal.Type().Field(i)
		if !hasMkvTag(fieldDesc) {
			continue
		}
		if err := inspectField(structVal.Field(i), fieldDesc); err != nil {
			return err
		}
	}
	return nil
}

func (r *Movie) writeField(xw *ixml.XmlWriter, field reflect.Value, tagName string) error {
	tagWriter := field.MethodByName("WriteTag")
	if tagWriter == (reflect.Value{}) {
		panic("field has no method \"WriteTag\"")
	}
	var ret []reflect.Value
	if len(tagName) < 1 {
		ret = tagWriter.Call([]reflect.Value{reflect.ValueOf(xw)})
	} else {
		ret = tagWriter.Call([]reflect.Value{reflect.ValueOf(xw), reflect.ValueOf(tagName)})
	}
	if ret == nil || len(ret) < 1 {
		panic("ret cannot be nil and must contain one return value")
	}
	if !ret[0].IsNil() {
		return ret[0].Interface().(error)
	}

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

	if err := r.writeFields(xw); err != nil {
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
