function openModal(id) {
  var modal = document.getElementById(id);
  if (modal) modal.hidden = false;
}

function closeModal(id) {
  var modal = document.getElementById(id);
  if (modal) modal.hidden = true;
}

if (window.location.pathname === "/") {
  document
    .querySelectorAll(".post-card .post-options, .post-card .modal")
    .forEach(function (controls) {
      controls.hidden = true;
    });
}

function showNotification(message, isError) {
  const notification = document.createElement("div");
  notification.className =
    "success-notification" + (isError ? " error-notification" : "");
  notification.setAttribute("role", isError ? "alert" : "status");
  notification.setAttribute("aria-live", "polite");
  notification.textContent = message;
  document.body.appendChild(notification);
  window.setTimeout(() => notification.remove(), 2000);
}

if (document.body.dataset.authenticated !== "true") {
  document
    .querySelectorAll(".reaction-button[data-post-id]")
    .forEach((button) => {
      button.disabled = true;
      button.title =
        button.dataset.reaction === "1"
          ? "Sign in to like this post"
          : "Sign in to dislike this post";
    });
}

document.addEventListener("click", async function (event) {
  const button = event.target.closest(".reaction-button[data-post-id]");
  if (!button || button.disabled) return;

  const post = button.closest(".post-card, .post-detail");
  const actions = button.closest(".post-card-actions");
  const status = post.querySelector(".reaction-status");
  if (actions.dataset.reacting === "true") return;

  actions.dataset.reacting = "true";
  actions.querySelectorAll(".reaction-button").forEach((reactionButton) => {
    reactionButton.disabled = true;
  });
  status.textContent = "";

  try {
    const response = await fetch(
      "/posts/" + encodeURIComponent(button.dataset.postId) + "/reaction",
      {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: new URLSearchParams({ reaction: button.dataset.reaction }),
      },
    );

    if (!response.ok) {
      throw new Error("Reaction request failed with status " + response.status);
    }

    const result = await response.json();
    const likeButton = actions.querySelector(
      '.reaction-button[data-reaction="1"]',
    );
    const dislikeButton = actions.querySelector(
      '.reaction-button[data-reaction="-1"]',
    );
    likeButton.querySelector(".reaction-count").textContent = result.like_count;
    dislikeButton.querySelector(".reaction-count").textContent =
      result.dislike_count;

    const liked = result.user_reaction === 1;
    const disliked = result.user_reaction === -1;
    likeButton.classList.toggle("liked", liked);
    dislikeButton.classList.toggle("disliked", disliked);
    likeButton.setAttribute("aria-pressed", String(liked));
    dislikeButton.setAttribute("aria-pressed", String(disliked));
  } catch (error) {
    console.error(error);
    status.textContent = "Could not update reaction. Please try again.";
  } finally {
    delete actions.dataset.reacting;
    const authenticated = document.body.dataset.authenticated === "true";
    actions.querySelectorAll(".reaction-button").forEach((reactionButton) => {
      reactionButton.disabled = !authenticated;
    });
  }
});

// Auth modal tab-switching: any element with [data-auth-switch="login"|"register"]
// closes both auth modals and opens the requested one.
document.addEventListener("DOMContentLoaded", function () {
  if (document.body.dataset.loginSuccess === "true") {
    const url = new URL(window.location.href);
    url.searchParams.delete("login_success");
    window.history.replaceState(null, "", url.pathname + url.search + url.hash);
    showNotification("Successfully logged in.", false);
  }

  const profilePictureNotification =
    document.body.dataset.profilePictureNotification;
  if (profilePictureNotification) {
    const url = new URL(window.location.href);
    url.searchParams.delete("profile_picture_status");
    window.history.replaceState(null, "", url.pathname + url.search + url.hash);
    showNotification(profilePictureNotification, false);
  }

  const loginError = document.body.dataset.loginError;
  if (loginError) {
    const error = document.getElementById("login-error");
    if (error) {
      error.textContent = loginError;
      error.hidden = false;
      openModal("login-modal");
    }
  }

  const registerError = document.body.dataset.registerError;
  if (registerError) {
    const error = document.getElementById("register-error");
    if (error) {
      error.textContent = registerError;
      error.hidden = false;
      openModal("register-modal");
    }
  }

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

  if (!input || !preview ) return;

  const uploadForm = input.form;
  uploadForm.addEventListener("submit", function (event) {
    const file = input.files && input.files[0];
    if (!file || file.size <= 5 * 1024 * 1024) return;

    event.preventDefault();
    event.stopPropagation();
    showNotification(
      "Failed to update profile picture. Image must be 5 MiB or smaller.",
      true,
    );
  });
  uploadForm.addEventListener("submit", async function (event) {
    if (uploadForm.dataset.confirmed !== "true") return;
    event.preventDefault();

    const username = uploadForm.elements.username.value.trim();
    const file = input.files && input.files[0];

    let response = await fetch("/users/username", {
      method: "POST",
      body: new URLSearchParams({ username }),
    });
    if (!response.ok) {
      showNotification(
        response.status === 409
          ? "That username is already taken."
          : "Username must be between 3 and 20 characters.",
        true,
      );
      return;
    }

    if (file) {
      const body = new FormData();
      body.append("profile_picture", file);
      response = await fetch("/users/profile_picture", { method: "POST", body });
      if (!response.ok) {
        showNotification("Failed to update profile picture.", true);
        return;
      }
    }

    location.reload();
  });
  let previewUrl;

  input.addEventListener("change", function () {
    const file = input.files && input.files[0];

    if (!file) {
      preview.replaceChildren();
      return;
    }

    if (previewUrl) URL.revokeObjectURL(previewUrl);

    previewUrl = URL.createObjectURL(file);

    const image = document.createElement("img");
    image.src = previewUrl;
    image.alt = "Selected profile picture";

    preview.replaceChildren(image);
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
