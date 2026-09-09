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
