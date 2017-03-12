package untappd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Beer ...
type Beer struct {
	ID          int      `json:"bid"`
	Name        string   `json:"beer_name"`
	Label       string   `json:"beer_label"`
	Ibu         int      `json:"beer_ibu"`
	Abv         float64  `json:"beer_abv"`
	Style       string   `json:"beer_style"`
	Description string   `json:"beer_description"`
	Brewery     *Brewery `json:"brewery,omitempty"`
}

// Brewery ...
type Brewery struct {
	Name string `json:"brewery_name"`
}

// Item ...
type Item struct {
	Beer    *Beer
	Brewery *Brewery
}

// Title ...
func (i *Item) Title() string {
	slug := func(s string) string {
		re := regexp.MustCompile("[^a-z0-9]+")
		return strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
	}

	u := fmt.Sprintf("https://untappd.com/b/%s/%d", slug(i.Beer.Name), i.Beer.ID)
	return fmt.Sprintf("<%s|%s>", u, fmt.Sprintf("%s %s", i.Brewery.Name, i.Beer.Name))
}

// Text ...
func (i *Item) Text() string {
	return fmt.Sprintf("%s | %d IBU | %0.0f%% ABV \n%s", i.Beer.Style, i.Beer.Ibu, i.Beer.Abv, i.Beer.Description)
}

// SearchResponse ...
type SearchResponse struct {
	Response struct {
		Beers struct {
			Items []*Item
		}
	}
}

// InfoResponse ...
type InfoResponse struct {
	Response struct {
		Beer *Beer `json:"beer"`
	} `json:"response"`
}

// Title ...
func (i *InfoResponse) Title() string {
	beer := i.Response.Beer

	// TODO: get slug from response
	slug := func(s string) string {
		re := regexp.MustCompile("[^a-z0-9]+")
		return strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
	}

	u := fmt.Sprintf("https://untappd.com/b/%s/%d", slug(beer.Name), beer.ID)
	return fmt.Sprintf("<%s|%s>", u, fmt.Sprintf("%s %s", beer.Brewery.Name, beer.Name))
}

// Text ...
func (i *InfoResponse) Text() string {
	beer := i.Response.Beer
	return fmt.Sprintf("%s | %d IBU | %0.0f%% ABV \n%s", beer.Style, beer.Ibu, beer.Abv, beer.Description)
}

type credentials struct {
	ClientID     string
	ClientSecret string
}

func getCredentials() (*credentials, error) {
	cid := os.Getenv("UNTAPPD_CLIENT_ID")
	if cid == "" {
		return nil, errors.New("missing environment variable: UNTAPPD_CLIENT_ID")
	}

	cs := os.Getenv("UNTAPPD_CLIENT_SECRET")
	if cs == "" {
		return nil, errors.New("missing environment variable: UNTAPPD_CLIENT_SECRET")
	}

	return &credentials{ClientID: cid, ClientSecret: cs}, nil
}

// Search ...
func Search(searchStr string) (*SearchResponse, error) {
	creds, err := getCredentials()
	if err != nil {
		return nil, err
	}

	u, err := url.Parse("https://api.untappd.com/v4/search/beer")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	q := u.Query()
	q.Set("client_id", creds.ClientID)
	q.Set("client_secret", creds.ClientSecret)
	q.Set("q", searchStr)
	u.RawQuery = q.Encode()

	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	var ur SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&ur); err != nil {
		return nil, err
	}

	return &ur, nil
}

// Info ...
func Info(id string) (*InfoResponse, error) {
	creds, err := getCredentials()
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(fmt.Sprintf("https://api.untappd.com/v4/beer/info/%s", id))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	q := u.Query()
	q.Set("client_id", creds.ClientID)
	q.Set("client_secret", creds.ClientSecret)
	q.Set("compact", "true")
	u.RawQuery = q.Encode()

	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	var ir InfoResponse
	if err := json.NewDecoder(res.Body).Decode(&ir); err != nil {
		return nil, err
	}

	return &ir, nil
}
