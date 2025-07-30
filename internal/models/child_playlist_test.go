package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildChildPlaylistName(t *testing.T) {
	tests := []struct {
		name              string
		basePlaylistName  string
		childPlaylistName string
		expectedResult    string
		expectedContains  []string
	}{
		{
			name:              "standard names",
			basePlaylistName:  "My Favorites",
			childPlaylistName: "High Energy",
			expectedResult:    "[My Favorites] > High Energy",
			expectedContains:  []string{"[My Favorites]", ">", "High Energy"},
		},
		{
			name:              "names with brackets and arrows",
			basePlaylistName:  "Base [with] brackets",
			childPlaylistName: "Child > with > arrows",
			expectedResult:    "[Base [with] brackets] > Child > with > arrows",
			expectedContains:  []string{"[Base [with] brackets]", ">", "Child > with > arrows"},
		},
		{
			name:              "empty child name",
			basePlaylistName:  "Base Playlist",
			childPlaylistName: "",
			expectedResult:    "[Base Playlist] > ",
			expectedContains:  []string{"[Base Playlist]", ">"},
		},
		{
			name:              "empty base name",
			basePlaylistName:  "",
			childPlaylistName: "Child Playlist",
			expectedResult:    "[] > Child Playlist",
			expectedContains:  []string{"[]", ">", "Child Playlist"},
		},
		{
			name:              "both names empty",
			basePlaylistName:  "",
			childPlaylistName: "",
			expectedResult:    "[] > ",
			expectedContains:  []string{"[]", ">"},
		},
		{
			name:              "names with unicode characters",
			basePlaylistName:  "Música Latina 🎵",
			childPlaylistName: "Reggaetón 💃",
			expectedResult:    "[Música Latina 🎵] > Reggaetón 💃",
			expectedContains:  []string{"[Música Latina 🎵]", ">", "Reggaetón 💃"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			result := BuildChildPlaylistName(tt.basePlaylistName, tt.childPlaylistName)

			// Verify exact result
			assert.Equal(tt.expectedResult, result)

			// Verify contains expected components
			for _, expectedPart := range tt.expectedContains {
				assert.Contains(result, expectedPart, "Result should contain: %s", expectedPart)
			}
		})
	}
}

func TestBuildChildPlaylistDescription(t *testing.T) {
	tests := []struct {
		name                string
		inputDescription    string
		expectedContains    []string
		expectedNotContains []string
		expectedLines       int
	}{
		{
			name:             "with user description",
			inputDescription: "This is my custom description for the child playlist",
			expectedContains: []string{
				"This is my custom description for the child playlist",
			},
			expectedLines: 5,
		},
		{
			name:             "empty description",
			inputDescription: "",
			expectedContains: []string{},
			expectedLines:    5, // Still 5 lines, but last line is empty
		},
		{
			name:             "multiline user description",
			inputDescription: "Line 1 of description\nLine 2 of description\nLine 3 of description",
			expectedContains: []string{
				"Line 1 of description",
				"Line 2 of description",
				"Line 3 of description",
			},
			expectedLines: 7, // 4 warning lines + 3 user description lines
		},
		{
			name:             "description with special characters",
			inputDescription: "Special chars: @#$%^&*()_+{}|:<>?[]\\;'\",./ and unicode: 🎵🎶💃🕺",
			expectedContains: []string{
				"Special chars: @#$%^&*()_+{}|:<>?[]\\;'\",./ and unicode: 🎵🎶💃🕺",
			},
			expectedLines: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			result := BuildChildPlaylistDescription(tt.inputDescription)

			// Verify the result is not empty
			assert.NotEmpty(result, "Result should never be empty")

			// Verify expected content is present
			for _, expectedContent := range tt.expectedContains {
				assert.Contains(result, expectedContent, "Result should contain: %s", expectedContent)
			}

			// Verify line count
			lines := strings.Split(result, "\n")
			assert.Len(lines, tt.expectedLines, "Should have expected number of lines")

			// Verify warning header is always present and at the beginning
			assert.Contains(lines[0], "****************************************************")
			assert.Contains(lines[1], "* PLAYLIST GENERATED AND MANAGED BY PlaylistRouter *")
			assert.Contains(lines[2], "*********** PLEASE DO NOT EDIT OR DELETE ***********")
			assert.Contains(lines[3], "****************************************************")

			// Verify user description is at the end (if provided)
			if tt.inputDescription != "" {
				// For multiline descriptions, the user content starts at line 4
				userContentLines := lines[4:]
				userContent := strings.Join(userContentLines, "\n")
				assert.Equal(tt.inputDescription, userContent)
			}
		})
	}
}
