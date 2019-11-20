package client

import (
	"fmt"
	"log"
	"os"
)

// replace uses file linking to safely replace dst with src.
func replace(oldfile, newfile string) error {
	// 1. Link src to a backup
	backup := oldfile + ".backup"
	if err := os.Link(oldfile, backup); err != nil {
		return fmt.Errorf("replace can't backup %s to %s because: %v", oldfile, backup, err)
	}

	// 2. Remove oldfile now that we've backed it up.
	if err := os.Remove(oldfile); err != nil {
		// Best effort to remove backup too.
		if err := os.Remove(backup); err != nil {
			log.Printf("Can't cleanup backup %s because %v\n", backup, err)
		}
		return fmt.Errorf("replace can't remove %s because %v", oldfile, err)
	}

	// 3. Put newfile in place of oldfile
	if err := os.Link(newfile, oldfile); err != nil {
		// Try to fix things.
		if err := os.Link(backup, oldfile); err != nil {
			log.Printf("Can't fix %s with  %s because %v\n", oldfile, backup, err)
		} else {
			if err := os.Remove(backup); err != nil {
				log.Printf("Can't cleanup restored backup %s because %v\n", backup, err)
			}
		}
		return fmt.Errorf("can't link %s to %s because: %v", oldfile, backup, err)
	}

	// 4. Cleanup
	if err := os.Remove(backup); err != nil {
		return fmt.Errorf("Can't cleanup unnecessary backup %s because %v\n", backup, err)
	}
	if err := os.Remove(newfile); err != nil {
		return fmt.Errorf("Can't remove newfile %s because %v\n", newfile, err)
	}
	return nil
}
