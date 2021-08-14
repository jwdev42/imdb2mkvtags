//This file is part of imdb2mkvtags ©2021 Jörg Walter

package schema

import (
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"html"
)

type Thing struct {
	AdditionalType            string `json:"additionalType"`
	AlternateName             string `json:"alternateName"`
	Description               string `json:"description"`
	DisambiguatingDescription string `json:"disambiguatingDescription"`
	Identifier                string `json:"identifier"`
	Image                     string `json:"image"`
	MainEntityOfPage          string `json:"mainEntityOfPage"`
	Name                      string `json:"name"`
	PotentialAction           string `json:"potentialAction"`
	SameAs                    string `json:"sameAs"`
	SubjectOf                 string `json:"subjectOf"`
	Type                      string `json:"@type"`
	Url                       string `json:"url"`
}

type Movie struct {
	Thing
	Actors        []Thing  `json:"actor"`
	ContentRating string   `json:"contentRating"`
	Context       string   `json:"@context"`
	Creators      []Thing  `json:"creator"`
	DatePublished string   `json:"datePublished"`
	Directors     []Thing  `json:"director"`
	Genres        []string `json:"genre"`
	Keywords      string   `json:"keywords"`
}

//Converts the imdb-imported json movie schema to imdb2mkvtags' internal data type.
//Text is HTML unescaped as a side effect.
func (r *Movie) Convert() *tags.Movie {
	//Naming convention:
	//Variables derived from the receiver have the prefix 's' if they can be confused
	//with variables derived from the struct tags.Movie

	movie := new(tags.Movie)

	if r.Actors != nil && len(r.Actors) > 0 {
		actors := make([]tags.Actor, 0, len(r.Actors))
		for _, sActor := range r.Actors {
			if len(sActor.Name) > 0 {
				actors = append(actors, tags.Actor{Name: html.UnescapeString(sActor.Name)})
			}
		}
		if len(actors) > 0 {
			movie.Actors = actors
		}
	}

	if len(r.ContentRating) > 0 {
		movie.ContentRating = make([]tags.MultiLingual, 1)
		movie.ContentRating[0].Text = html.UnescapeString(r.ContentRating)
	}

	if len(r.DatePublished) > 0 {
		movie.ReleaseDate = tags.UniLingual(html.UnescapeString(r.DatePublished))
	}

	if len(r.Description) > 0 {
		movie.Synopses = make([]tags.MultiLingual, 1)
		movie.Synopses[0].Text = html.UnescapeString(r.Description)
	}

	if r.Directors != nil && len(r.Directors) > 0 {
		directors := make([]tags.UniLingual, 0, len(r.Directors))
		for _, sDirector := range r.Directors {
			if len(sDirector.Name) > 0 {
				directors = append(directors, tags.UniLingual(html.UnescapeString(sDirector.Name)))
			}
		}
		if len(directors) > 0 {
			movie.Directors = directors
		}
	}
	if r.Genres != nil && len(r.Genres) > 0 {
		genres := make([]tags.MultiLingual, 0, len(r.Genres))
		for _, sGenre := range r.Genres {
			genres = append(genres, tags.MultiLingual{Text: html.UnescapeString(sGenre)})
		}
		movie.Genres = genres
	}

	if len(r.Keywords) > 0 {
		movie.Keywords = tags.UniLingual(r.Keywords)
	}

	if len(r.Name) > 0 {
		movie.Titles = make([]tags.MultiLingual, 1)
		movie.Titles[0].Text = html.UnescapeString(r.Name)
	}

	return movie
}