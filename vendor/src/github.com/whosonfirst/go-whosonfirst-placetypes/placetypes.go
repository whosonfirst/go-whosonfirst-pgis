package placetypes

import (
	"encoding/json"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-placetypes/spec"
	"strconv"
)

type WOFPlacetypeSpec map[string]WOFPlacetype

type WOFPlacetypes struct {
	spec WOFPlacetypeSpec
}

type WOFPlacetypeName struct {
	Lang string `json:"language"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type WOFPlacetypeAltNames map[string][]string

type WOFPlacetype struct {
	Id     int64   `json:"id"`
	Name   string  `json:"name"`
	Role   string  `json:"role"`
	Parent []int64 `json:"parent"`
	// AltNames []WOFPlacetypeAltNames		`json:"names"`
}

func Init() (*WOFPlacetypes, error) {

	places := placetypes.Spec()

	var spec WOFPlacetypeSpec
	err := json.Unmarshal([]byte(places), &spec)

	if err != nil {
		return nil, err
	}

	placetypes := WOFPlacetypes{
		spec: spec,
	}

	return &placetypes, nil
}

func (pt *WOFPlacetypes) GetPlacetypeByName(name string) (*WOFPlacetype, error) {

	for str_id, pt := range pt.spec {

		if pt.Name == name {

			id, _ := strconv.Atoi(str_id)
			pt.Id = int64(id)

			return &pt, nil
		}
	}

	return nil, errors.New("Invalid placetype")
}

/*
func (sp *WOFPlacetypeSpec) Common() []string {
	return sp.WithRole("common")
}

func (sp *WOFPlacetypeSpec) CommonOptional() []string {
	return sp.WithRole("common_optional")
}

func (sp *WOFPlacetypeSpec) Optional() []string {
	return sp.WithRole("optional")
}

func (sp *WOFPlacetypeSpec) WithRole(role string) []string {

	places := make([]string, 0)

	for id, placetype := range sp {
		if placetype.Role != role {
		   continue
		}

		places = append(places, role)
	}

	return places
}

func (sp *WOFPlacetypeSpec) WithRoles(roles []string) []string {

	places := make([]string, 0)

	for _, role := range roles {
	    places = append(places, sp.WithRole(role))
	}

	return places
}

func IsValidRole(role string) bool {

	return false
}
*/
