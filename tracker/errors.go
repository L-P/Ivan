package tracker

import "fmt"

type errInvalidZone struct {
	zone, zoneItem int
}

func (e errInvalidZone) Error() string {
	return fmt.Sprintf("invalid zoneKP/itemKP (%d/%d), must be [1-9]", e.zone, e.zoneItem)
}

type errNoDefinition struct {
	zone, zoneItem int
}

func (e errNoDefinition) Error() string {
	return fmt.Sprintf("no definition for item at %d/%d", e.zone, e.zoneItem)
}

type errInvalidDefinition struct {
	zone, zoneItem int
	name           string
}

func (e errInvalidDefinition) Error() string {
	return fmt.Sprintf(`bad item name "%s" at %d/%d"`, e.name, e.zone, e.zoneItem)
}
