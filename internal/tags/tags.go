//This file is part of imdb2mkvtags ©2021 Jörg Walter

package tags

import (
	"encoding/xml"
	ixml "github.com/jwdev42/imdb2mkvtags/internal/xml"
	"io"
	"strings"
)

type Actor struct {
	Name      string
	Character string
}

func (r *Actor) WriteTag(xw *ixml.XmlWriter) error {
	var fc func(*ixml.XmlWriter) error
	if r.Character != "" {
		character := make([][2]string, 0, 2)
		character = append(character, [2]string{"Name", "CHARACTER"})
		character = append(character, [2]string{"String", r.Character})
		fc = ixml.GenWriteTagWithSubtags("Simple", character, nil)
	}
	actor := make([][2]string, 0, 2)
	actor = append(actor, [2]string{"Name", "ACTOR"})
	actor = append(actor, [2]string{"String", r.Name})
	return ixml.WriteTagWithSubtags(xw, "Simple", actor, fc)
}

type UniLingual string

func (r UniLingual) WriteTag(xw *ixml.XmlWriter, name string) error {
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
	subtags := make([][2]string, 0, 3)
	subtags = append(subtags, [2]string{"Name", name})
	subtags = append(subtags, [2]string{"String", r.Text})
	if r.Lang != "" {
		subtags = append(subtags, [2]string{"TagLanguageIETF", r.Lang})
	}
	return ixml.WriteTagWithSubtags(xw, "Simple", subtags, nil)
}

type Movie struct {
	Actors        []Actor
	ContentRating []MultiLingual
	Directors     []UniLingual
	Genres        []MultiLingual
	Imdb          UniLingual
	Keywords      []UniLingual
	Producers     []UniLingual
	ReleaseDate   UniLingual
	Synopses      []MultiLingual
	Titles        []MultiLingual
	Writers       []UniLingual
}

func (r *Movie) WriteTag(xw *ixml.XmlWriter) error {

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

	for _, actor := range r.Actors {
		if err := actor.WriteTag(xw); err != nil {
			return err
		}
	}

	for _, cr := range r.ContentRating {
		if err := cr.WriteTag(xw, "LAW_RATING"); err != nil {
			return err
		}
	}

	for _, director := range r.Directors {
		if err := director.WriteTag(xw, "DIRECTOR"); err != nil {
			return err
		}
	}

	for _, genre := range r.Genres {
		if err := genre.WriteTag(xw, "GENRE"); err != nil {
			return err
		}
	}

	if len(r.Imdb) > 0 {
		if err := r.Imdb.WriteTag(xw, "IMDB"); err != nil {
			return err
		}
	}

	if len(r.Keywords) > 0 {
		buf := new(strings.Builder)
		for i := 0; i < len(r.Keywords); i++ {
			buf.WriteString(string(r.Keywords[i]))
			if i < len(r.Keywords)-1 {
				buf.WriteByte(',')
			}
		}
		if err := UniLingual(buf.String()).WriteTag(xw, "KEYWORDS"); err != nil {
			return err
		}
	}

	for _, producer := range r.Producers {
		if err := producer.WriteTag(xw, "PRODUCER"); err != nil {
			return err
		}
	}

	if len(r.ReleaseDate) > 0 {
		if err := r.ReleaseDate.WriteTag(xw, "DATE_RELEASED"); err != nil {
			return err
		}
	}

	for _, synopsis := range r.Synopses {
		if err := synopsis.WriteTag(xw, "SYNOPSIS"); err != nil {
			return err
		}
	}

	for _, title := range r.Titles {
		if err := title.WriteTag(xw, "TITLE"); err != nil {
			return err
		}
	}

	for _, writer := range r.Writers {
		if err := writer.WriteTag(xw, "WRITTEN_BY"); err != nil {
			return err
		}
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
	return nil
}
