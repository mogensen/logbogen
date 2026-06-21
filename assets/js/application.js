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
    const defaultIcon = L.icon({
        iconUrl: '/images/map/pin.png',
        iconSize: [36, 36],
        iconAnchor: [18, 36],
        popupAnchor: [0, -36],
    });


    $(() => {
        if ($('#participants').length == 0) {
            return;
        }

        var data = loadUsers();
        data.then(data => {
            $('#participants').select2({
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
    if ($('#activity-form').length == 0) {
        return;
    }

    async function loadCategories() {
        const categoryGroup = $("#category-group");
        const currentCategory = categoryGroup.data('current') || "climbing";
        const pillMode = categoryGroup.hasClass('act-edit-pillrow');

        try {
            const response = await fetch('/activities/categories');
            if (!response.ok) {
                throw new Error('Failed to fetch categories');
            }

            const categories = await response.json();
            categoryGroup.empty();

            categories.forEach(category => {
                const selected = category.ID === currentCategory;
                if (pillMode) {
                    const pillHtml = `<label class="act-edit-pill${selected ? ' sel' : ''}">
                        <img src="/images/categories/${category.ID}.png" alt="${category.Name}">
                        ${category.Name}
                        <input type="radio" name="category" value="${category.ID}" ${selected ? 'checked' : ''} autocomplete="off" style="display:none">
                    </label>`;
                    categoryGroup.append(pillHtml);
                } else {
                    const radioId = `category-${category.ID}`;
                    const radioHtml = `
                    <div class="category-radio-parent">
                        <input type="radio" class="btn-check category-radio" name="category" id="${radioId}" value="${category.ID}"
                            autocomplete="off" ${selected ? 'checked' : ''}>
                        <label class="btn btn-outline-primary category-label" for="${radioId}">
                            <img src="/images/categories/${category.ID}.png" alt="${category.Name}" class="category-image">
                            <br />
                            <span>${category.Name}</span>
                        </label>
                    </div>
                    `;
                    categoryGroup.append(radioHtml);
                }
            });

            updateTypeOptions();
        } catch (error) {
            console.error('Error loading categories:', error);
        }
    }

    async function updateTypeOptions() {
        const category = $("input[name='category']:checked").val();
        const typeGroup = $("#activity-type-group");
        // Read data-current only once on initial load; after that use the live checked radio.
        // removeAttr ensures jQuery can't re-read the stale HTML attribute on subsequent calls.
        const currentType = typeGroup.data('current') || $("input[name='type']:checked").val();
        const pillMode = typeGroup.hasClass('act-edit-pillrow');

        if (category == undefined) {
            return;
        }

        try {
            const response = await fetch(`/activities/types?category=${category}`);
            if (!response.ok) {
                throw new Error('Failed to fetch activity types');
            }

            const types = await response.json();
            typeGroup.empty();
            typeGroup.removeData('current');
            typeGroup.removeAttr('data-current');

            types.forEach(type => {
                const selected = type.ID === currentType;
                if (pillMode) {
                    const pillHtml = `<label class="act-edit-pill${selected ? ' sel' : ''}">
                        <img src="/images/activities/${type.ID}.png" alt="${type.Name}">
                        ${type.Name}
                        <input type="radio" name="type" value="${type.ID}" ${selected ? 'checked' : ''} autocomplete="off" style="display:none" required>
                    </label>`;
                    typeGroup.append(pillHtml);
                } else {
                    const radioId = `type-${type.ID}`;
                    const radioHtml = `
                    <div class="type-radio-parent">
                        <input type="radio" class="btn-check type-radio form-check-input" name="type" id="${radioId}" value="${type.ID}"
                            autocomplete="off" ${selected ? 'checked' : ''} required>
                        <label class="btn btn-outline-primary type-label form-check-label" for="${radioId}">
                            <img src="/images/activities/${type.ID}.png" alt="${type.Name}" class="category-image">
                            <br />
                            <span>${type.Name}</span>
                        </label>
                    </div>
                    `;
                    typeGroup.append(radioHtml);
                }
            });

            // Sync header icon with whichever type is selected after render
            if (pillMode) {
                const $icon = $('#act-edit-type-icon');
                if ($icon.length) {
                    if (category === 'other') {
                        $icon.attr('src', '/images/activities/other.png').css('opacity', '1');
                    } else if (currentType && types.some(t => t.ID === currentType)) {
                        $icon.attr('src', `/images/activities/${currentType}.png`).css('opacity', '1');
                    } else {
                        $icon.css('opacity', '0');
                    }
                }
            }

            updateOtherType();
        } catch (error) {
            console.error('Error updating activity types:', error);
        }
    }

    $(document).on('click', '.act-edit-pillrow .act-edit-pill', function() {
        const $pill = $(this);
        $pill.closest('.act-edit-pillrow').find('.act-edit-pill').removeClass('sel');
        $pill.addClass('sel');
        const $radio = $pill.find('input[type=radio]');
        $radio.prop('checked', true).trigger('change');
        if ($radio.attr('name') === 'type') {
            $('#act-edit-type-icon')
                .attr('src', `/images/activities/${$radio.val()}.png`)
                .css('opacity', '1');
        }
    });

    // Update types when category changes
    $(document).on('change', 'input[name="category"]', updateTypeOptions);

    function updateOtherType() {
        if ($("input[name='category']:checked").val() === "other") {
            $("#activity-type-group").parent().slideUp();
            $("#activity-othertype-group").slideDown();
            // Set input as required
            $("#activity-othertype").prop('required', true);
        } else {
            $("#activity-type-group").parent().slideDown();
            $("#activity-othertype-group").slideUp();
            // Set input as not required
            $("#activity-othertype").prop('required', false);
        }
    }

    // Update other type field when type changes
    $(document).on('change', 'input[name="type"]', updateOtherType);

    // Hide errors when a valid selection is made
    $(document).on('change', 'input[name="type"]', function() {
        $('#type-error').hide();
    });
    $(document).on('input', '#activity-othertype', function() {
        if ($(this).val().trim()) $('#othertype-error').hide();
    });

    // Validate type/othertype on submit
    $('#activity-form').on('submit', function(e) {
        const category = $("input[name='category']:checked").val();
        const typeSelected = $("input[name='type']:checked").val();
        const otherTypeVal = $('#activity-othertype').val().trim();

        if (category === 'other' && !otherTypeVal) {
            e.preventDefault();
            $('#othertype-error').show();
            $('#activity-othertype')[0]?.scrollIntoView({ behavior: 'smooth', block: 'center' });
        } else if (category !== 'other' && !typeSelected) {
            e.preventDefault();
            $('#type-error').show();
            $('#type-error')[0]?.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
    });

    // Initial load of categories
    loadCategories();
    // Initial update
    updateTypeOptions();

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

// Example starter JavaScript for disabling form submissions if there are invalid fields
(function () {
    'use strict'
  
    // Fetch all the forms we want to apply custom Bootstrap validation styles to
    var forms = document.querySelectorAll('.needs-validation')
  
    // Loop over them and prevent submission
    Array.prototype.slice.call(forms)
      .forEach(function (form) {
        form.addEventListener('submit', function (event) {
          if (!form.checkValidity()) {
            event.preventDefault()
            event.stopPropagation()
          }
  
          form.classList.add('was-validated')
        }, false)
      })
  })()