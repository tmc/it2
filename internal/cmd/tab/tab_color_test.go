package tab

import (
	"testing"
)

func TestParseRGB(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantRed   float64
		wantGreen float64
		wantBlue  float64
		wantErr   bool
	}{
		{
			name:      "red color",
			input:     "255,0,0",
			wantRed:   1.0,
			wantGreen: 0.0,
			wantBlue:  0.0,
			wantErr:   false,
		},
		{
			name:      "green color",
			input:     "0,255,0",
			wantRed:   0.0,
			wantGreen: 1.0,
			wantBlue:  0.0,
			wantErr:   false,
		},
		{
			name:      "blue color",
			input:     "0,0,255",
			wantRed:   0.0,
			wantGreen: 0.0,
			wantBlue:  1.0,
			wantErr:   false,
		},
		{
			name:      "white color",
			input:     "255,255,255",
			wantRed:   1.0,
			wantGreen: 1.0,
			wantBlue:  1.0,
			wantErr:   false,
		},
		{
			name:      "black color",
			input:     "0,0,0",
			wantRed:   0.0,
			wantGreen: 0.0,
			wantBlue:  0.0,
			wantErr:   false,
		},
		{
			name:      "gray color",
			input:     "128,128,128",
			wantRed:   128.0 / 255.0,
			wantGreen: 128.0 / 255.0,
			wantBlue:  128.0 / 255.0,
			wantErr:   false,
		},
		{
			name:      "with spaces",
			input:     " 255 , 128 , 64 ",
			wantRed:   1.0,
			wantGreen: 128.0 / 255.0,
			wantBlue:  64.0 / 255.0,
			wantErr:   false,
		},
		{
			name:    "invalid - too few components",
			input:   "255,0",
			wantErr: true,
		},
		{
			name:    "invalid - too many components",
			input:   "255,0,0,0",
			wantErr: true,
		},
		{
			name:    "invalid - non-numeric",
			input:   "red,green,blue",
			wantErr: true,
		},
		{
			name:    "invalid - out of range high",
			input:   "256,0,0",
			wantErr: true,
		},
		{
			name:    "invalid - out of range low",
			input:   "-1,0,0",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			red, green, blue, err := parseRGB(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRGB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				const epsilon = 0.001
				if abs(red-tt.wantRed) > epsilon || abs(green-tt.wantGreen) > epsilon || abs(blue-tt.wantBlue) > epsilon {
					t.Errorf("parseRGB() = (%.3f, %.3f, %.3f), want (%.3f, %.3f, %.3f)",
						red, green, blue, tt.wantRed, tt.wantGreen, tt.wantBlue)
				}
			}
		})
	}
}

func TestParseHex(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantRed   float64
		wantGreen float64
		wantBlue  float64
		wantErr   bool
	}{
		{
			name:      "red with hash",
			input:     "#FF0000",
			wantRed:   1.0,
			wantGreen: 0.0,
			wantBlue:  0.0,
			wantErr:   false,
		},
		{
			name:      "red without hash",
			input:     "FF0000",
			wantRed:   1.0,
			wantGreen: 0.0,
			wantBlue:  0.0,
			wantErr:   false,
		},
		{
			name:      "green",
			input:     "#00FF00",
			wantRed:   0.0,
			wantGreen: 1.0,
			wantBlue:  0.0,
			wantErr:   false,
		},
		{
			name:      "blue",
			input:     "#0000FF",
			wantRed:   0.0,
			wantGreen: 0.0,
			wantBlue:  1.0,
			wantErr:   false,
		},
		{
			name:      "white",
			input:     "#FFFFFF",
			wantRed:   1.0,
			wantGreen: 1.0,
			wantBlue:  1.0,
			wantErr:   false,
		},
		{
			name:      "black",
			input:     "#000000",
			wantRed:   0.0,
			wantGreen: 0.0,
			wantBlue:  0.0,
			wantErr:   false,
		},
		{
			name:      "gray",
			input:     "#808080",
			wantRed:   128.0 / 255.0,
			wantGreen: 128.0 / 255.0,
			wantBlue:  128.0 / 255.0,
			wantErr:   false,
		},
		{
			name:      "lowercase hex",
			input:     "#ff00ff",
			wantRed:   1.0,
			wantGreen: 0.0,
			wantBlue:  1.0,
			wantErr:   false,
		},
		{
			name:      "mixed case hex",
			input:     "#FfAa00",
			wantRed:   1.0,
			wantGreen: 170.0 / 255.0,
			wantBlue:  0.0,
			wantErr:   false,
		},
		{
			name:    "invalid - too short",
			input:   "#FF00",
			wantErr: true,
		},
		{
			name:    "invalid - too long",
			input:   "#FF000000",
			wantErr: true,
		},
		{
			name:    "invalid - non-hex characters",
			input:   "#GGHHII",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid - just hash",
			input:   "#",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			red, green, blue, err := parseHex(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				const epsilon = 0.001
				if abs(red-tt.wantRed) > epsilon || abs(green-tt.wantGreen) > epsilon || abs(blue-tt.wantBlue) > epsilon {
					t.Errorf("parseHex() = (%.3f, %.3f, %.3f), want (%.3f, %.3f, %.3f)",
						red, green, blue, tt.wantRed, tt.wantGreen, tt.wantBlue)
				}
			}
		})
	}
}

func TestRGBToHex(t *testing.T) {
	tests := []struct {
		name  string
		red   float64
		green float64
		blue  float64
		want  string
	}{
		{
			name:  "red",
			red:   1.0,
			green: 0.0,
			blue:  0.0,
			want:  "#FF0000",
		},
		{
			name:  "green",
			red:   0.0,
			green: 1.0,
			blue:  0.0,
			want:  "#00FF00",
		},
		{
			name:  "blue",
			red:   0.0,
			green: 0.0,
			blue:  1.0,
			want:  "#0000FF",
		},
		{
			name:  "white",
			red:   1.0,
			green: 1.0,
			blue:  1.0,
			want:  "#FFFFFF",
		},
		{
			name:  "black",
			red:   0.0,
			green: 0.0,
			blue:  0.0,
			want:  "#000000",
		},
		{
			name:  "gray",
			red:   128.0 / 255.0,
			green: 128.0 / 255.0,
			blue:  128.0 / 255.0,
			want:  "#808080",
		},
		{
			name:  "custom color",
			red:   200.0 / 255.0,
			green: 100.0 / 255.0,
			blue:  50.0 / 255.0,
			want:  "#C86432",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rgbToHex(tt.red, tt.green, tt.blue)
			if got != tt.want {
				t.Errorf("rgbToHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRGBHexRoundTrip(t *testing.T) {
	// Test that parsing and converting back yields the original value
	tests := []struct {
		name string
		hex  string
	}{
		{"red", "#FF0000"},
		{"green", "#00FF00"},
		{"blue", "#0000FF"},
		{"white", "#FFFFFF"},
		{"black", "#000000"},
		{"gray", "#808080"},
		{"custom", "#A1B2C3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse hex to RGB
			red, green, blue, err := parseHex(tt.hex)
			if err != nil {
				t.Fatalf("parseHex() error = %v", err)
			}

			// Convert back to hex
			got := rgbToHex(red, green, blue)
			if got != tt.hex {
				t.Errorf("round trip: parseHex(%s) -> rgbToHex() = %v, want %v", tt.hex, got, tt.hex)
			}
		})
	}
}

func TestRGBParsing(t *testing.T) {
	// Test that parsing RGB and converting to hex works correctly
	tests := []struct {
		name    string
		rgb     string
		wantHex string
	}{
		{"red", "255,0,0", "#FF0000"},
		{"green", "0,255,0", "#00FF00"},
		{"blue", "0,0,255", "#0000FF"},
		{"white", "255,255,255", "#FFFFFF"},
		{"black", "0,0,0", "#000000"},
		{"gray", "128,128,128", "#808080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			red, green, blue, err := parseRGB(tt.rgb)
			if err != nil {
				t.Fatalf("parseRGB() error = %v", err)
			}

			got := rgbToHex(red, green, blue)
			if got != tt.wantHex {
				t.Errorf("parseRGB(%s) -> rgbToHex() = %v, want %v", tt.rgb, got, tt.wantHex)
			}
		})
	}
}

// Helper function
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
