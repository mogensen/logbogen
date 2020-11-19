require("expose-loader?$!expose-loader?jQuery!jquery");
require("bootstrap/dist/js/bootstrap.bundle.js");
require("@fortawesome/fontawesome-free/js/all.js");
require("leaflet/dist/leaflet-src.js");
require("leaflet-easybutton/src/easy-button.js");
require("select2/dist/js/select2.full.js");

$(() => {

    if ($('#user-show-map').length == 0) {
        return;
    }

    var map = new L.Map('user-show-map');
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
            id: 'user-show-map',
        }).addTo(map);

    var markerArray = [];
    markers.forEach(function(row){

        var icon = new L.icon({
            iconUrl: '/assets/images/climbing/' + row.type.toLowerCase() + '.svg',
            iconSize: [35, 35],
            iconAnchor: [22, 35]
          });

        var m = L.marker(
            new L.LatLng(row.lat,row.lng), {
                icon: icon,
                title: row.type,
            });

        markerArray.push(m);
        m.addTo(map);
    });

    var group = L.featureGroup(markerArray).addTo(map);
    map.fitBounds(group.getBounds());

});

