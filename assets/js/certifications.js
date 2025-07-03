
$(() => {
    if ($('#certification-form').length == 0) {
        return;
    }

    async function loadCategories() {
        console.log("loadCategories");
        const categoryGroup = $("#category-group");
        const currentCategory = categoryGroup.data('current') || "firstaid";
        
        try {
            const response = await fetch('/certifications/categories');
            if (!response.ok) {
                throw new Error('Failed to fetch categories');
            }
            
            const categories = await response.json();
            
            // Clear existing radio buttons
            categoryGroup.empty();
            
            // Add new radio buttons for each category
            categories.forEach(category => {
                const radioId = `category-${category.ID}`;
                const radioHtml = `
                <div class="category-radio-parent">
                    <input type="radio" class="btn-check category-radio" name="category" id="${radioId}" value="${category.ID}"
                        autocomplete="off" ${category.ID === currentCategory ? 'checked' : ''}>
                    <label class="btn btn-outline-primary category-label" for="${radioId}">
                        <img src="/images/categories/${category.ID}.png" alt="${category.Name}" class="category-image">
                        <br />
                        <span>${category.Name}</span>
                    </label>
                </div>
                `;
                categoryGroup.append(radioHtml);
            });
            
            // Initial update of types
            updateTypeOptions();
        } catch (error) {
            console.error('Error loading categories:', error);
        }
    }

    async function updateTypeOptions() {
        const category = $("input[name='category']:checked").val();
        const typeGroup = $("#type");
        const currentType = typeGroup.data('current') || typeGroup.val();

        if (category == undefined) {
            console.log("undefined category")
            return;
        }

        try {
            const response = await fetch(`/certifications/types?category=${category}`);
            if (!response.ok) {
                throw new Error('Failed to fetch activity types');
            }
            
            const types = await response.json();
            
            // Clear existing radio buttons
            typeGroup.empty();
            
            // Add new radio buttons with images
            types.forEach(type => {
                const opion = `<option value="${type.ID}" ${type.ID === currentType ? 'selected' : ''} >${type.Name}</option>`;
                typeGroup.append(opion);
            });
            
            // Update other type field visibility
            updateOtherType();
        } catch (error) {
            console.error('Error updating activity types:', error);
        }
    }

    // Update types when category changes
    $(document).on('change', 'input[name="category"]', updateTypeOptions);

    function updateOtherType() {
        const category = $("input[name='category']:checked").val();
        if (category === "other") {
            $("#type").parent().slideUp();
            $("#othertype-group").slideDown();
            // Set input as required
            $("#activity-othertype").prop('required', true);
        } else {
            $("#type").parent().slideDown();
            $("#othertype-group").slideUp();
            // Set input as not required
            $("#activity-othertype").prop('required', false);
        }
    }

    // Update other type field when type changes
    $(document).on('change', 'input[name="type"]', updateOtherType);

    // Initial load of categories
    loadCategories();
    // Initial update
    updateTypeOptions();

});
