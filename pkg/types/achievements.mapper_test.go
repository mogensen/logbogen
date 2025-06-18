package types

import (
	"reflect"
	"testing"
)

func TestAchievements(t *testing.T) {
	tests := []struct {
		name               string
		activities         []*Activity
		wantedAchievements []Achievement
	}{
		{
			name:       "No activities",
			activities: []*Activity{},
			wantedAchievements: func() []Achievement {
				var achs []Achievement
				for _, at := range AllActivityTypes {
					achs = append(achs, Achievement{
						Type:  at,
						Level: 0,
					})
				}
				return achs
			}(),
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
			wantedAchievements: func() []Achievement {
				var achs []Achievement
				for _, at := range AllActivityTypes {
					if at == AllActivityTypes[0] {
						achs = append(achs, Achievement{
							Type:  at,
							Level: 1,
						})
					} else {
						achs = append(achs, Achievement{
							Type:  at,
							Level: 0,
						})
					}
				}
				return achs
			}(),
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
			wantedAchievements: func() []Achievement {
				var achs []Achievement
				for i, at := range AllActivityTypes {
					if i >= 0 && i <= 4 {
						achs = append(achs, Achievement{
							Type:  at,
							Level: 1,
						})
					} else {
						achs = append(achs, Achievement{
							Type:  at,
							Level: 0,
						})
					}
				}
				return achs
			}(),
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
			wantedAchievements: func() []Achievement {
				var achs []Achievement
				for i, at := range AllActivityTypes {
					switch i {
					case 0:
						achs = append(achs, Achievement{
							Type:  at,
							Level: 1,
						})
					case 1:
						achs = append(achs, Achievement{
							Type:  at,
							Level: 2,
						})
					case 2:
						achs = append(achs, Achievement{
							Type:  at,
							Level: 1,
						})
					default:
						achs = append(achs, Achievement{
							Type:  at,
							Level: 0,
						})
					}
				}
				return achs
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAchievements := Achievements(tt.activities); !reflect.DeepEqual(gotAchievements, tt.wantedAchievements) {
				t.Errorf("Achievements() = %v, want %v", gotAchievements, tt.wantedAchievements)
			}
		})
	}
}
