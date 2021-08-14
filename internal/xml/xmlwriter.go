//This file is part of imdb2mkvtags ©2021 Jörg Walter

package xml

import (
	"encoding/xml"
	"io"
)

type endStack struct {
	elements int
	top      *endStackElem
}

type endStackElem struct {
	pred *endStackElem
	e    xml.EndElement
}

func (r *endStack) push(e xml.EndElement) {
	se := new(endStackElem)
	se.e = e
	se.pred = r.top
	r.top = se
	r.elements += 1
}

func (r *endStack) pop() xml.EndElement {
	if r.elements < 1 {
		panic("cannot pop an empty stack")
	}
	e := r.top.e
	r.top = r.top.pred
	r.elements -= 1
	return e
}

type XmlWriter struct {
	*xml.Encoder
	s endStack
}

func NewXmlWriter(w io.Writer) *XmlWriter {
	return &XmlWriter{
		Encoder: xml.NewEncoder(w),
		s:       endStack{},
	}
}

func (r *XmlWriter) EncodeToken(t xml.Token) error {
	if se, ok := t.(xml.StartElement); ok {
		if err := r.Encoder.EncodeToken(se); err != nil {
			return err
		}
		r.s.push(se.End())
	} else if _, ok := t.(xml.EndElement); ok {
		panic("XmlWriter: EncodeToken does not support EndElement, use CloseElement() instead")
	} else {
		return r.Encoder.EncodeToken(t)
	}
	return nil
}

func (r *XmlWriter) EncodeTokens(tokens ...xml.Token) error {
	for _, t := range tokens {
		if err := r.EncodeToken(t); err != nil {
			return err
		}
	}
	return nil
}

func (r *XmlWriter) CloseElement() error {
	ee := r.s.pop()
	if err := r.Encoder.EncodeToken(ee); err != nil {
		r.s.push(ee)
		return err
	}
	return nil
}

func (r *XmlWriter) CloseAllElements() error {
	for i := r.s.elements; i > 0; i-- {
		if err := r.CloseElement(); err != nil {
			return err
		}
	}
	return nil
}

func (r *XmlWriter) OpenElements() int {
	return r.s.elements
}

//Xml-escapes b, then writes its content to a text node
func (r *XmlWriter) WriteText(b []byte) error {
	return r.EncodeToken(xml.CharData(b))
}
