package dict

import (
	"encoding/xml"
	"io"
	"os"
	"strings"
)

// ParsedDictionary holds raw information extracted from OpenCorpora XML.
type ParsedDictionary struct {
	Lexemes   map[string][]WordForm
	Links     []Link
	Grammemes []Grammeme
	Version   string
	Revision  string
}

// WordForm represents word with its tag.
type WordForm struct {
	Word string
	Tag  string
}

// Link represents dictionary link.
type Link struct {
	From string
	To   string
	Type string
}

// Grammeme represents grammeme description.
type Grammeme struct {
	Name        string
	Parent      string
	Alias       string
	Description string
}

// getDictionaryInfo returns version and revision from XML file.
func getDictionaryInfo(filename string) (string, string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", err
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "dictionary" {
			ver := attr(se.Attr, "version")
			rev := attr(se.Attr, "revision")
			return ver, rev, nil
		}
	}
	return "", "", nil
}

func attr(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// ParseOpencorporaXML parses XML dictionary file.
func ParseOpencorporaXML(filename string) (*ParsedDictionary, error) {
	data := &ParsedDictionary{
		Lexemes:   map[string][]WordForm{},
		Links:     []Link{},
		Grammemes: []Grammeme{},
	}
	ver, rev, _ := getDictionaryInfo(filename)
	data.Version = ver
	data.Revision = rev

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var xdict struct {
		XMLName   xml.Name `xml:"dictionary"`
		Grammemes []struct {
			Name        string `xml:"name"`
			Parent      string `xml:"parent,attr"`
			Alias       string `xml:"alias"`
			Description string `xml:"description"`
		} `xml:"grammeme"`
		Lemmata []struct {
			ID string `xml:"id,attr"`
			L  struct {
				Gs []struct {
					V string `xml:"v,attr"`
				} `xml:"g"`
			} `xml:"l"`
			Fs []struct {
				T  string `xml:"t,attr"`
				Gs []struct {
					V string `xml:"v,attr"`
				} `xml:"g"`
			} `xml:"f"`
		} `xml:"lemma"`
		Links []struct {
			From string `xml:"from,attr"`
			To   string `xml:"to,attr"`
			Type string `xml:"type,attr"`
		} `xml:"link"`
	}
	if err := xml.Unmarshal(content, &xdict); err != nil {
		return nil, err
	}

	for _, g := range xdict.Grammemes {
		data.Grammemes = append(data.Grammemes, Grammeme{
			Name:        g.Name,
			Parent:      g.Parent,
			Alias:       g.Alias,
			Description: g.Description,
		})
	}

	for _, l := range xdict.Lemmata {
		base := joinGrams(l.L.Gs)
		forms := []WordForm{}
		for _, f := range l.Fs {
			gram := joinGrams(f.Gs)
			tag := strings.TrimSpace(base + " " + gram)
			forms = append(forms, WordForm{Word: strings.ToLower(f.T), Tag: tag})
		}
		data.Lexemes[l.ID] = forms
	}

	for _, ln := range xdict.Links {
		data.Links = append(data.Links, Link{From: ln.From, To: ln.To, Type: ln.Type})
	}

	return data, nil
}

func joinGrams(gs []struct {
	V string `xml:"v,attr"`
}) string {
	parts := make([]string, 0, len(gs))
	for _, g := range gs {
		parts = append(parts, g.V)
	}
	return strings.Join(parts, ",")
}
