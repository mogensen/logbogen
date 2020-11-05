package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"logbogen/models"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/markbates/goth"
)

type facebookProfile struct {
	Data struct {
		Height       int    `json:"height"`
		IsSilhouette bool   `json:"is_silhouette"`
		URL          string `json:"url"`
		Width        int    `json:"width"`
	} `json:"data"`
}

func TryUpdateImage(c buffalo.Context, gu goth.User, u *models.User) error {

	// Get the profile image json from FB
	resp, err := http.Get(fmt.Sprintf("https://graph.facebook.com/v8.0/%s/picture?height=300&width=300&redirect=false&access_token=%s", gu.UserID, gu.AccessToken))
	if err != nil {
		return err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Unmarshal profile image json to struct
	profile := facebookProfile{}
	err = json.Unmarshal(body, &profile)
	if err != nil {
		return err
	}

	// If FB is only sending a silhouette, we just give up
	if profile.Data.IsSilhouette {
		return nil
	}

	// Download actual profile image
	resp, err = http.Get(profile.Data.URL)
	if err != nil {
		return err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Update image in database
	pImage := &models.UsersImage{
		UserID: u.ID,
	}

	if u.Image != nil {
		pImage = u.Image
	}

	pImage.UserID = u.ID

	pImage.ImageData = []byte(base64.StdEncoding.EncodeToString(body))
	tx := c.Value("tx").(*pop.Connection)

	if err = tx.Save(pImage); err != nil {
		return err
	}
	return nil
}
