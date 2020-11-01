require("expose-loader?$!expose-loader?jQuery!jquery");
require("bootstrap/dist/js/bootstrap.bundle.js");
require("@fortawesome/fontawesome-free/js/all.js");
require("leaflet/dist/leaflet-src.js");
require("leaflet-easybutton/src/easy-button.js");

import img from 'leaflet/dist/images/marker-icon.png';

$(() => {
    const zoomLevel = 15;
    const defaultIcon = new L.icon({
        iconUrl: img,
        iconAnchor: [12.5, 41],
    });
    $(() => {

        var lat = $('#activity-show-map').data("lat");
        var lng = $('#activity-show-map').data("lng");

        var map = L.map('activity-show-map').setView([lat, lng], zoomLevel);
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
            id: 'activity-form-map',
        }).addTo(map);

        var marker = L.marker([lat, lng], { icon: defaultIcon }).addTo(map);
    })

    function addMapPicker() {

        var mapCenter = [22, 87];
        var map = L.map('activity-form-map', { center: mapCenter, zoom: 3 });
        map.setView(new L.LatLng(56.05, +9.85), 7);

        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
            id: 'activity-form-map',
        }).addTo(map);

        function locate(control){
            control.state("loading");
            control._map.on('locationfound', function(e){
              this.setView(e.latlng, 16);
              control.state('loaded');
            });
            control._map.on('locationerror', function(){
              control.state('error');
            });
            control._map.locate()
          }

        L.easyButton({
            states:[
              {
                stateName: 'unloaded',
                icon: 'fa-location-arrow',
                title: 'load image',
                onClick: locate,
              }, {
                stateName: 'loading',
                icon: 'fa-spinner fa-spin'
              }, {
                stateName: 'loaded',
                icon: 'fa-location-arrow',
                onClick: locate,
              }, {
                stateName: 'error',
                icon: 'fa-frown-o',
                title: 'location not found'
              }
            ]
          }).addTo(map);

        var marker = L.marker(mapCenter, { icon: defaultIcon }).addTo(map);
        function updateMarker(lat, lng) {
            marker.setLatLng([lat, lng]);
            return false;
        };

        map.on('click', function (e) {
            $('#climbingactivity-Lat').val(e.latlng.lat);
            $('#climbingactivity-Lng').val(e.latlng.lng);
            updateMarker(e.latlng.lat, e.latlng.lng);
        });

        var lat = $('#activity-form-map').data("lat");
        var lng = $('#activity-form-map').data("lng");

        if (lat != "" && lng != "") {
            updateMarker(lat, lng);
            map.setView(new L.LatLng(lat, lng), zoomLevel);
        }

    }

    addMapPicker();

});
