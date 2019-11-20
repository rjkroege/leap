package server

import (
	"fmt"
	"log"
)

func (t *Server) Ping(arg string, result *string) error {
	log.Printf("ping. client sez %s\n", arg)

	if arg == "" {
		return fmt.Errorf("rude ping without greeting")
	}
	*result = fmt.Sprintf("%s back to you!", arg)
	return nil
}
