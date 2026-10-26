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

document.addEventListener("DOMContentLoaded", function () {
  const confirmModal = document.getElementById("confirm-modal");
  const confirmMessage = document.querySelector("[data-confirm-message]");
  const confirmSubmit = document.querySelector("[data-confirm-submit]");
  let pendingForm = null;

  if (!confirmModal || !confirmMessage || !confirmSubmit) return;

  document.addEventListener("submit", function (event) {
    const form = event.target.closest("form[data-confirm]");
    if (!form) return;

    if (form.dataset.confirmed === "true") {
      delete form.dataset.confirmed;
      return;
    }

    event.preventDefault();
    pendingForm = form;
    confirmMessage.textContent = form.dataset.confirm;
    openModal("confirm-modal");
  });

  document.querySelectorAll("[data-confirm-cancel]").forEach(function (button) {
    button.addEventListener("click", function () {
      pendingForm = null;
      closeModal("confirm-modal");
    });
  });

  confirmSubmit.addEventListener("click", function () {
    if (!pendingForm) return;

    const form = pendingForm;
    pendingForm = null;
    form.dataset.confirmed = "true";
    closeModal("confirm-modal");
    form.requestSubmit();
  });
});

// Post options menus (<details class="post-options">): only one open at a
// time, closed by clicking outside or pressing Escape.
document.addEventListener("click", function (event) {
  document.querySelectorAll(".post-options[open]").forEach(function (menu) {
    if (!menu.contains(event.target)) menu.open = false;
  });
});

document.addEventListener("keydown", function (event) {
  if (event.key !== "Escape") return;
  document.querySelectorAll(".post-options[open]").forEach(function (menu) {
    menu.open = false;
  });
});

// Search clear button: if a search is active, go back to the full feed;
// otherwise just empty the box without reloading.
document.addEventListener("click", function (event) {
  const button = event.target.closest(".search-clear");
  if (!button) return;

  if (button.hasAttribute("data-active-search")) {
    window.location.href = "/";
    return;
  }

  const input = button.form.querySelector(".search-input");
  input.value = "";
  input.focus();
});

// After posting a reply the page reloads at #comment-N. If that comment is
// inside a collapsed "View N replies" dropdown, open it (and any
// dropdowns around it) so the browser can scroll to it.
document.addEventListener("DOMContentLoaded", function () {
  if (!location.hash) return;
  const target = document.getElementById(location.hash.slice(1));
  if (!target) return;

  let details = target.closest("details.more-replies");
  if (!details) return;
  while (details) {
    details.open = true;
    details = details.parentElement.closest("details.more-replies");
  }
  target.scrollIntoView();
});

// Comment and reply boxes: Enter sends, Shift+Enter adds a new line.
// Skipped mid-composition so IME users can confirm characters with Enter.
document.addEventListener("keydown", function (event) {
  if (event.key !== "Enter" || event.shiftKey || event.isComposing) return;

  const textarea = event.target.closest(
    ".comment-form textarea, .reply-form textarea",
  );
  if (!textarea) return;

  event.preventDefault();
  textarea.form.requestSubmit();
});

// Category picker: checkboxes can't use `required` (that would mean "check
// ALL of them"), so "pick at least one" is enforced manually via
// setCustomValidity on the first checkbox in the group. That plugs into the
// same native validation the browser already uses for `required` fields, so
// it blocks submission and shows the same native tooltip.
document.querySelectorAll(".category-picker").forEach((picker) => {
  const boxes = picker.querySelectorAll('input[type="checkbox"]');

  function validate() {
    const anyChecked = Array.prototype.some.call(boxes, (box) => box.checked);
    boxes[0].setCustomValidity(
      anyChecked ? "" : "Please select at least one category.",
    );
  }

  boxes.forEach((box) => box.addEventListener("change", validate));
  validate();
});

//script for profile image preview
document.addEventListener("DOMContentLoaded", function () {
  const input = document.getElementById("profile-picture-input");
  const preview = document.getElementById("avatar-preview");
  const saveButton = document.getElementById("save-picture");

  if (!input || !preview || !saveButton) return;

  let previewUrl;

  input.addEventListener("change", function () {
    const file = input.files && input.files[0];

    if (!file) {
      preview.replaceChildren();
      saveButton.hidden = true;
      return;
    }

    if (previewUrl) URL.revokeObjectURL(previewUrl);

    previewUrl = URL.createObjectURL(file);

    const image = document.createElement("img");
    image.src = previewUrl;
    image.alt = "Selected profile picture";

    preview.replaceChildren(image);
    saveButton.hidden = false;
  });
});
// Comment delete: send the request in the background and remove the
// comment (and its replies) from the page instead of reloading.
document.addEventListener("submit", function (event) {
  const form = event.target.closest(".comment-delete-form");
  if (!form) return;

  event.preventDefault();

  fetch(form.action, { method: "POST" })
    .then(function (response) {
      if (!response.ok) throw new Error(response.status);
      form.closest(".comment-thread").remove();
    })
    .catch(function () {
      alert("Couldn't delete the comment. Please try again.");
    });
});
