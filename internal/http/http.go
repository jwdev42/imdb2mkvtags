//This file is part of imdb2mkvtags ©2021 Jörg Walter

package http

import (
	"io"
	"net/http"
)

func Body(url string, dest io.Writer) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, err := dest.Write(buf[:n])
			if err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

	}
	return nil
}
