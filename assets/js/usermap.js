require("expose-loader?$!expose-loader?jQuery!jquery");
require("bootstrap/dist/js/bootstrap.bundle.js");
require("@fortawesome/fontawesome-free/js/all.js");
require("leaflet/dist/leaflet-src.js");
require("leaflet-easybutton/src/easy-button.js");

$(() => {

    if ($('#user-show-map').length == 0) {
        return;
    }

    var map = new L.Map('user-show-map').setView([56, 10], 5);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
        id: 'user-show-map',
    }).addTo(map);

    var lang = getCookie("lang");
    var markerArray = [];

    markers.forEach(function (row) {
        console.log(row)

        var icon = new L.icon({
            iconUrl: '/assets/images/climbing/' + row.type.toLowerCase() + '.svg',
            iconSize: [35, 35],
            iconAnchor: [22, 35]
        });

        var m = L.marker(
            new L.LatLng(row.lat, row.lng), {
            icon: icon,
            title: markerTitle(row),
            win_url: "/climbingactivities/" + row.id,
        });

        markerArray.push(m);
        m.addTo(map);
        m.on('click', onClick);
    });

    function onClick(e) {
        window.location = this.options.win_url;
    }

    function markerTitle(row, lang) {
        var dt = new Date(row.date);
        var month = dt.toLocaleString(lang, { month: 'long' });
        var year = dt.toLocaleString(lang, { year: 'numeric' });
        return row.type.charAt(0).toUpperCase() + row.type.slice(1).toLowerCase() + " climbing " + month + " " + year;
    }

    var group = L.featureGroup(markerArray).addTo(map);
    map.fitBounds(group.getBounds());


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

