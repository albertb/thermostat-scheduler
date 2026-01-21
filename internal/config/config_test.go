
package config

import (
	"strings"
	"testing"
)

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{
			name: "valid config",
			config: `
username: user
password: password
peak_events_url: "https://example.com"
normal_program:
  sunday: &default_program
    morning: { time: 7h, heat: 21, cool: 24 }
    day:     { time: 9h, heat: 21, cool: 24 }
    evening: { time: 16h, heat: 21, cool: 24 }
    night:   { time: 21h, heat: 20, cool: 25 }
  monday: *default_program
  tuesday: *default_program
  wednesday: *default_program
  thursday: *default_program
  friday: *default_program
  saturday: *default_program
peak_program:
  pre_heat_duration: 1h
  pre_heat_temp_offset: 1
  peak_temp_offset: -1
`,
			wantErr: false,
		},
		{
			name: "invalid peak_events_url",
			config: `
username: user
password: password
peak_events_url: "invalid-url"
`,
			wantErr: true,
		},
		{
			name: "missing peak_events_url",
			config: `
username: user
password: password
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ReadConfig(strings.NewReader(tt.config))
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
