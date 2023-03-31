package general

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
)

func (c *commands) NewIDPair() error {
	short := shortid.New()
	long, err := shortid.ToUUID(short)
	if err != nil {
		return err
	}
	fmt.Printf("%s %s\n", long, short)
	return nil
}

func (c *commands) NewShortID() error {
	shortID := shortid.New()
	fmt.Printf("%s\n", shortID)
	return nil
}

func (c *commands) ToShortID(id string) error {
	shortID, err := shortid.ParseString(id)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", shortID)
	return nil
}

func (c *commands) ToLongID(id string) error {
	longID, err := shortid.ToUUID(id)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", longID)
	return nil
}
