//This file is part of imdb2mkvtags ©2021 Jörg Walter

package xml

import "encoding/xml"

// Returns a StartElement that has no attributes and resides in the default namespace
func NewStartElementSimple(name string) xml.StartElement {
	return xml.StartElement{Name: xml.Name{Local: name}}
}

func WriteTagWithSubtags(xw *XmlWriter, tag string, subtags [][2]string, post func(*XmlWriter) error) error {
	f := GenWriteTagWithSubtags(tag, subtags, post)
	return f(xw)
}

func GenWriteTagWithSubtags(tag string, subtags [][2]string, post func(xw *XmlWriter) error) func(xw *XmlWriter) error {
	return func(xw *XmlWriter) error {
		//Open tag
		if err := xw.EncodeToken(NewStartElementSimple(tag)); err != nil {
			return err
		}
		for _, subtag := range subtags {
			//Open subtag
			if err := xw.EncodeToken(NewStartElementSimple(subtag[0])); err != nil {
				return err
			}
			//Write Text
			if err := xw.WriteText([]byte(subtag[1])); err != nil {
				return err
			}
			//Close subtag
			if err := xw.CloseElement(); err != nil {
				return err
			}
		}
		//Exec post function
		if post != nil {
			if err := post(xw); err != nil {
				return err
			}
		}
		//Close tag
		if err := xw.CloseElement(); err != nil {
			return err
		}
		return nil
	}
}
