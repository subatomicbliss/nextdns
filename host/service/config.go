package service

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type ConfigStorer interface {
	SaveConfig(map[string]ConfigEntry) error
	LoadConfig(map[string]ConfigEntry) error
}

type ConfigEntry interface {
	Set(string) error
	String() string
}

type ConfigListEntry interface {
	ConfigEntry
	Strings() []string
}

type ConfigValue struct {
	Value *string
}

func (e ConfigValue) Set(v string) error {
	*e.Value = v
	return nil
}

func (e ConfigValue) String() string {
	if e.Value == nil {
		return ""
	}
	return *e.Value
}

type ConfigFlag struct {
	Value *bool
}

func (e ConfigFlag) Set(v string) error {
	switch v {
	case "yes", "true", "1":
		*e.Value = true
	case "no", "false", "0", "":
		*e.Value = false
	default:
		return fmt.Errorf("%v: invalid bool value", v)
	}
	return nil
}

func (e ConfigFlag) String() string {
	if e.Value != nil && *e.Value {
		return "true"
	} else {
		return "false"
	}
}

type ConfigDuration struct {
	Value *time.Duration
}

func (e ConfigDuration) Set(v string) error {
	d, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	*e.Value = d
	return nil
}

func (e ConfigDuration) String() string {
	if e.Value == nil {
		return ""
	}
	return e.Value.String()
}

type ConfigFileStorer struct {
	File string
}

func (s ConfigFileStorer) SaveConfig(c map[string]ConfigEntry) error {
	f, err := os.OpenFile(s.File, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	for name, entry := range c {
		if entry, ok := entry.(ConfigListEntry); ok {
			for _, value := range entry.Strings() {
				fmt.Fprintf(f, "%s %s\n", name, value)
			}
			continue
		}
		fmt.Fprintf(f, "%s %s\n", name, entry.String())
	}
	return nil
}

func (s ConfigFileStorer) LoadConfig(c map[string]ConfigEntry) error {
	f, err := os.Open(s.File)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name := line
		value := ""
		if idx := strings.IndexByte(line, ' '); idx != -1 {
			name = line[:idx]
			value = strings.TrimSpace(line[idx+1:])
		}
		if entry := c[name]; entry != nil {
			if err := entry.Set(value); err != nil {
				return err
			}
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return nil
}