// archivements.service_test.go

package services

import (
	"reflect"
	"testing"

	"github.com/mogensen/logbook/pkg/types"
)

func TestAchievements(t *testing.T) {
	tests := []struct {
		name               string
		activities         []*types.Activity
		wantedAchievements []types.Achievement
	}{
		{
			name:       "No activities",
			activities: []*types.Activity{},
			wantedAchievements: func() []types.Achievement {
				var achs []types.Achievement
				for _, at := range types.AllActivityTypes {
					achs = append(achs, types.Achievement{
						Type:  at,
						Level: 0,
					})
				}
				return achs
			}(),
		},
		{
			name: "Single activity type",
			activities: []*types.Activity{
				{Type: types.AllActivityTypes[0]},
				{Type: types.AllActivityTypes[0]},
				{Type: types.AllActivityTypes[0]},
				{Type: types.AllActivityTypes[0]},
				{Type: types.AllActivityTypes[0]},
			},
			wantedAchievements: func() []types.Achievement {
				var achs []types.Achievement
				for _, at := range types.AllActivityTypes {
					if at == types.AllActivityTypes[0] {
						achs = append(achs, types.Achievement{
							Type:  at,
							Level: 1,
						})
					} else {
						achs = append(achs, types.Achievement{
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
			activities: []*types.Activity{
				{Type: types.AllActivityTypes[0]},
				{Type: types.AllActivityTypes[1]},
				{Type: types.AllActivityTypes[2]},
				{Type: types.AllActivityTypes[3]},
				{Type: types.AllActivityTypes[4]},
			},
			wantedAchievements: func() []types.Achievement {
				var achs []types.Achievement
				for i, at := range types.AllActivityTypes {
					if i >= 0 && i <= 4 {
						achs = append(achs, types.Achievement{
							Type:  at,
							Level: 1,
						})
					} else {
						achs = append(achs, types.Achievement{
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
			activities: []*types.Activity{
				{Type: types.AllActivityTypes[0]},
				{Type: types.AllActivityTypes[0]},
				{Type: types.AllActivityTypes[1]},
				{Type: types.AllActivityTypes[1]},
				{Type: types.AllActivityTypes[1]},
				{Type: types.AllActivityTypes[1]},
				{Type: types.AllActivityTypes[1]},
				{Type: types.AllActivityTypes[1]},
				{Type: types.AllActivityTypes[2]},
			},
			wantedAchievements: func() []types.Achievement {
				var achs []types.Achievement
				for i, at := range types.AllActivityTypes {
					switch i {
					case 0:
						achs = append(achs, types.Achievement{
							Type:  at,
							Level: 1,
						})
					case 1:
						achs = append(achs, types.Achievement{
							Type:  at,
							Level: 2,
						})
					case 2:
						achs = append(achs, types.Achievement{
							Type:  at,
							Level: 1,
						})
					default:
						achs = append(achs, types.Achievement{
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
