require("expose-loader?$!expose-loader?jQuery!jquery");
require("bootstrap/dist/js/bootstrap.bundle.js");
require("@fortawesome/fontawesome-free/js/all.js");
require("leaflet/dist/leaflet-src.js");
require("leaflet-easybutton/src/easy-button.js");
require("select2/dist/js/select2.full.js");


import img from 'leaflet/dist/images/marker-icon.png';

$(() => {
    $.fn.select2.defaults.set( "theme", "bootstrap4" );

    const zoomLevel = 15;
    const defaultIcon = new L.icon({
        iconUrl: img,
        iconAnchor: [12.5, 41],
    });


    $(() => {
        $('#participant-select').select2({
        minimumInputLength: 3 // only start searching when the user has input 3 or more characters
      });
    })

    $(() => {
        if ($('#activity-show-map').length == 0) {
            return;
        }

        var lat = $('#activity-show-map').data("lat");
        var lng = $('#activity-show-map').data("lng");

        var map = L.map('activity-show-map').setView([lat, lng], zoomLevel);
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
            id: 'activity-form-map',
        }).addTo(map);

        L.marker([lat, lng], { icon: defaultIcon }).addTo(map);
    })

    function addMapPicker() {
        if ($('#activity-form-map').length == 0) {
            return;
        }

        var mapCenter = [22, 87];
        var map = L.map('activity-form-map', { center: mapCenter, zoom: 3 });
        map.setView(new L.LatLng(56.05, +9.85), 7);

        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
            id: 'activity-form-map',
        }).addTo(map);

        function locate(control) {
            control.state("loading");
            control._map.on('locationfound', function (e) {
                this.setView(e.latlng, 16);
                control.state('loaded');
            });
            control._map.on('locationerror', function () {
                control.state('error');
            });
            control._map.locate()
        }

        L.easyButton({
            states: [
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

$(() => {
    function updateOther() {
        if ($(this).val() == "OTHER") {
            $("#climbingactivity-OtherType").parent(".form-group").slideDown();
        } else {
            $("#climbingactivity-OtherType").parent(".form-group").slideUp();
        }
    }

    $("#climbingactivity-Type").on('change', updateOther);

    if ($("#climbingactivity-Type").val() == "OTHER") {
        $("#climbingactivity-OtherType").parent(".form-group").sjow();
    } else {
        $("#climbingactivity-OtherType").parent(".form-group").hide();
    }
});
