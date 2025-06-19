package types

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAchievements(t *testing.T) {
	tests := []struct {
		name               string
		activities         []*Activity
		wantedAchievements []Achievement
	}{
		{
			name:               "No activities",
			activities:         []*Activity{},
			wantedAchievements: []Achievement{},
		},
		{
			name: "Single activity type",
			activities: []*Activity{
				{Type: AllActivityTypes[0]},
				{Type: AllActivityTypes[0]},
				{Type: AllActivityTypes[0]},
				{Type: AllActivityTypes[0]},
				{Type: AllActivityTypes[0]},
			},
			wantedAchievements: []Achievement{
				{
					Type:  AllActivityTypes[0],
					Level: 1,
				},
			},
		},
		{
			name: "Multiple activity types",
			activities: []*Activity{
				{Type: AllActivityTypes[0]},
				{Type: AllActivityTypes[1]},
				{Type: AllActivityTypes[2]},
				{Type: AllActivityTypes[3]},
				{Type: AllActivityTypes[4]},
			},
			wantedAchievements: []Achievement{
				{Type: AllActivityTypes[0], Level: 1},
				{Type: AllActivityTypes[1], Level: 1},
				{Type: AllActivityTypes[2], Level: 1},
				{Type: AllActivityTypes[3], Level: 1},
				{Type: AllActivityTypes[4], Level: 1},
			},
		},
		{
			name: "Different counts per type",
			activities: []*Activity{
				{Type: AllActivityTypes[0]},
				{Type: AllActivityTypes[0]},
				{Type: AllActivityTypes[1]},
				{Type: AllActivityTypes[1]},
				{Type: AllActivityTypes[1]},
				{Type: AllActivityTypes[1]},
				{Type: AllActivityTypes[1]},
				{Type: AllActivityTypes[1]},
				{Type: AllActivityTypes[2]},
			},
			wantedAchievements: []Achievement{
				{Type: AllActivityTypes[0], Level: 1},
				{Type: AllActivityTypes[1], Level: 2},
				{Type: AllActivityTypes[2], Level: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAchievements := Achievements(tt.activities)
			if len(gotAchievements) == 0 && len(tt.wantedAchievements) == 0 {
				// Both are empty, treat as equal
				return
			}
			if !cmp.Equal(gotAchievements, tt.wantedAchievements) {
				t.Errorf("Achievements() = %v, want %v", gotAchievements, tt.wantedAchievements)
			}
		})
	}
}
