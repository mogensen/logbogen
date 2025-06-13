// require("expose-loader?$!expose-loader?jQuery!jquery");
// require("bootstrap/dist/js/bootstrap.bundle.js");
// require("@fortawesome/fontawesome-free/js/all.js");
// require("leaflet/dist/leaflet-src.js");
// require("leaflet-easybutton/src/easy-button.js");
// require("select2/dist/js/select2.full.js");

async function loadUsers() {
    try {
        const response = await fetch('/users/list', {
            headers: {
              'Accept': 'application/json'
            }});
        if (!response.ok) {
            throw new Error(`Response status: ${response.status}`);
        }

        return await fetch('/users/list', {
            headers: {
              'Accept': 'application/json'
            }})
        .then (response => response.json())
        .then(data => data.map(function (item) {
            return {
                id: item.id,
                text: item.name,
            };
        }))
        .catch(function(error){console.log(error);});
    } catch (error) {
        console.error(error.message);
        return [];
    }
}


$(() => {
    $.fn.select2.defaults.set("theme", "default");

    const zoomLevel = 12;
    const defaultIcon = new L.icon({
        iconUrl: "https://www.ippc.int/static/leaflet/images/marker-icon.png",
        iconAnchor: [12.5, 41],
    });


    $(() => {
        if ($('#activity-participants').length == 0) {
            return;
        }

        var data = loadUsers();
        data.then(data => {
            $('#activity-participants').select2({
                minimumInputLength: 3,
                data: data,
                placeholder: "Andre deltagere...",
            });
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

        // Add geocoder control
        L.Control.geocoder({
            defaultMarkGeocode: false,
            placeholder: 'Search for a location...',
            errorMessage: 'Nothing found.',
            showResultIcons: true,
            expand: 'touch',
            position: 'topleft'  // Position it on the left side
        }).on('markgeocode', function(e) {
            var bbox = e.geocode.bbox;
            var poly = L.polygon([
                bbox.getSouthEast(),
                bbox.getNorthEast(),
                bbox.getNorthWest(),
                bbox.getSouthWest()
            ]);
            map.fitBounds(poly.getBounds());

            // Update the marker position and form fields
            var center = poly.getBounds().getCenter();
            $('#activity-lat').val(center.lat);
            $('#activity-lng').val(center.lng);
            updateMarker(center.lat, center.lng);
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
            $('#activity-lat').val(e.latlng.lat);
            $('#activity-lng').val(e.latlng.lng);
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
    if ($('.category-radio').length == 0) {
        return;
    }

    async function updateTypeOptions() {
        const category = $("input[name='category']:checked").attr('id');
        console.log(category);
        const typeSelect = $("#activity-type");
        const currentType = typeSelect.val();  // Store current type before clearing
        
        try {
            const response = await fetch(`/activities/types?category=${category}`);
            if (!response.ok) {
                throw new Error('Failed to fetch activity types');
            }
            
            const types = await response.json();
            
            // Clear existing options
            typeSelect.find("option").remove();
            
            // Add new options
            Object.entries(types).forEach(([value, name]) => {
                typeSelect.append(`<option value="${value}">${name}</option>`);
            });
            
            // If the current type exists in the new options, keep it selected
            if (Object.keys(types).includes(currentType)) {
                typeSelect.val(currentType);
            } else if (currentType !== "other") {
                // If current type doesn't exist in new options, select the first one
                typeSelect.val(Object.keys(types)[0]);
            }
        } catch (error) {
            console.error('Error updating activity types:', error);
        }
    }

    // Update types when category changes
    $("input[name='category']").on('change', updateTypeOptions);

    function updateOther() {
        if ($("#activity-type").val() == "other") {
            $("#activity-othertype").parent(".form-group").slideDown();
        } else {
            $("#activity-othertype").parent(".form-group").slideUp();
        }
    }

    $("#activity-type").on('change', updateOther);

    // Initial update
    updateTypeOptions();
    updateOther();
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