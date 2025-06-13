// require("expose-loader?$!expose-loader?jQuery!jquery");
// require("bootstrap/dist/js/bootstrap.bundle.js");
// require("@fortawesome/fontawesome-free/js/all.js");
// require("leaflet/dist/leaflet-src.js");
// require("leaflet-easybutton/src/easy-button.js");

$(() => {

    if ($('#user-show-map').length == 0) {
        return;
    }

    var map = new L.Map('user-show-map').setView([56, 10], 7);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
        id: 'user-show-map',
    }).addTo(map);

    var lang = getCookie("lang");
    var markerArray = [];

    fetch('/activities/list', {
        headers: {
            'Accept': 'application/json',
        }
    })
        .then(response => {
            if (response.ok) {
                return response.json(); // Parse the response data as JSON
            } else {
                throw new Error('API request failed');
            }
        })
        .then(data => {
            // Process the response data here
            data.forEach(function (row) {

                var icon = new L.icon({
                    iconUrl: '/images/activities/' + row.type.ID + '.svg',
                    iconSize: [35, 35],
                    iconAnchor: [22, 35]
                });

                var m = L.marker(
                    new L.LatLng(row.lat, row.lng), {
                    icon: icon,
                    title: markerTitle(row),
                    win_url: "/activities/" + row.id,
                });

                markerArray.push(m);
                m.addTo(map);
                m.on('click', onClick);
            })

            if (markerArray.length <= 1) {
                map.setView([markerArray[0].getLatLng().lat, markerArray[0].getLatLng().lng], 10);
            } else {
                var group = L.featureGroup(markerArray).addTo(map);
                map.fitBounds(group.getBounds().pad(0.1));
            }

        })
        .catch(error => {
            // Handle any errors here
            console.error(error); // Example: Logging the error to the console
        });

    function onClick(e) {
        window.location = this.options.win_url;
    }

    function markerTitle(row, lang) {
        var dt = new Date(row.date);
        var month = dt.toLocaleString(lang, { month: 'long' });
        var year = dt.toLocaleString(lang, { year: 'numeric' });
        return row.type.Name + " " + month + " " + year;
    }

    function getCookie(name) {
        var nameEQ = name + "=";
        var ca = document.cookie.split(';');
        for (var i = 0; i < ca.length; i++) {
            var c = ca[i];
            while (c.charAt(0) == ' ') c = c.substring(1, c.length);
            if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length, c.length);
        }
        return null;
    }

});

