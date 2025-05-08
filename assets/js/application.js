// require("expose-loader?$!expose-loader?jQuery!jquery");
// require("bootstrap/dist/js/bootstrap.bundle.js");
// require("@fortawesome/fontawesome-free/js/all.js");
// require("leaflet/dist/leaflet-src.js");
// require("leaflet-easybutton/src/easy-button.js");
// require("select2/dist/js/select2.full.js");


$(() => {
    $.fn.select2.defaults.set("theme", "classic");

    const zoomLevel = 12;
    const defaultIcon = new L.icon({
        iconUrl: "https://www.ippc.int/static/leaflet/images/marker-icon.png",
        iconAnchor: [12.5, 41],
    });


    $(() => {
        $('#climbingactivity-participants').select2({
            minimumInputLength: 3,
            ajax: {
                url: '/users/list',
                dataType: 'json',
                // Additional AJAX parameters go here; see the end of this chapter for the full code of this example
                processResults: function (data) {
                    // Transform: [{"id":1,"name":"Frederik Mogensen","email":"frede@server-1.dk"},{"id":2,"name":"Tine Stenum","email":"tine@mail.com"}]
                    // into: {  "results": [{"id":1,"text":"Frederik Mogensen"},{"id":2,"text":"Tine Stenum"}], {  "pagination": {    "more": true  } }

                    res = data.map(function (item) {
                        return {
                            id: item.id,
                            text: item.name + " (" + item.email + ")",
                        };
                    });
                    console.log(res);

                    // Transforms the top-level key of the response object from 'items' to 'results'
                    return {
                        results: res
                    };
                },
            }
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
            $('#climbingactivity-lat').val(e.latlng.lat);
            $('#climbingactivity-lng').val(e.latlng.lng);
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
        console.log("updateOther");
        console.log($(this).val());
        if ($(this).val() == "other") {
            $("#climbingactivity-othertype").parent(".form-group").slideDown();
        } else {
            $("#climbingactivity-othertype").parent(".form-group").slideUp();
        }
    }

    $("#climbingactivity-type").on('change', updateOther);

    if ($("#climbingactivity-type").val() == "other") {
        $("#climbingactivity-othertype").parent(".form-group").show();
    } else {
        $("#climbingactivity-othertype").parent(".form-group").hide();
    }
});


$(() => {
    // Bind click to OK button within popup
    $('#confirm-delete').on('click', '.btn-ok', function (e) {
        console.log("confirm-delete");
        console.log($(this));

        var $modalDiv = $(e.delegateTarget);
        var id = $(this).data('recordId');

        $modalDiv.addClass('loading');
        $.post('/activities/' + id + '/delete').then(function () {
            location.href = '/activities/list';
        });
    });

    // Bind to modal opening to set necessary data properties to be used to make request
    $('#confirm-delete').on('show.bs.modal', function (e) {
        console.log($(e.relatedTarget))
        var data = $(e.relatedTarget).data();
        $('.title', this).text(data.recordTitle);
        $('.btn-ok', this).data('recordId', data.recordId);
    });
});