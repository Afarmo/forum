function openModal(id) {
    var modal = document.getElementById(id);
    if (modal) modal.hidden = false;
}

function closeModal(id) {
    var modal = document.getElementById(id);
    if (modal) modal.hidden = true;
}

// Auth modal tab-switching: any element with [data-auth-switch="login"|"register"]
// closes both auth modals and opens the requested one.
document.addEventListener("DOMContentLoaded", function () {
    document.querySelectorAll("[data-auth-switch]").forEach(function (el) {
        el.addEventListener("click", function () {
            closeModal("login-modal");
            closeModal("register-modal");
            openModal(el.dataset.authSwitch + "-modal");
        });
    });
});
// Category picker: checkboxes can't use `required` (that would mean "check
// ALL of them"), so "pick at least one" is enforced manually via
// setCustomValidity on the first checkbox in the group. That plugs into the
// same native validation the browser already uses for `required` fields, so
// it blocks submission and shows the same native tooltip.
document.querySelectorAll(".category-picker").forEach(picker => {
    const boxes = picker.querySelectorAll('input[type="checkbox"]');

    function validate() {
        const anyChecked = Array.prototype.some.call(boxes, box => box.checked);
        boxes[0].setCustomValidity(anyChecked ? "" : "Please select at least one category.");
    }

    boxes.forEach(box => box.addEventListener("change", validate));
    validate();
});

document.querySelectorAll("[data-image-source]").forEach(tab => {
    tab.addEventListener("click", () => {
        document.querySelectorAll("[data-image-source]").forEach(t => t.classList.remove("active"));
        tab.classList.add("active");
        document.querySelectorAll("[data-source-panel]").forEach(panel => {
            panel.hidden = panel.dataset.sourcePanel !== tab.dataset.imageSource;
        });
    });
});
