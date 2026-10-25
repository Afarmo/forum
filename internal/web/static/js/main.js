function openModal(id) {
  var modal = document.getElementById(id);
  if (modal) modal.hidden = false;
}

function closeModal(id) {
  var modal = document.getElementById(id);
  if (modal) modal.hidden = true;
}

if (document.body.dataset.authenticated !== "true") {
  document.querySelectorAll(".post-card .reaction-button").forEach((button) => {
    button.disabled = true;
    button.title =
      button.dataset.reaction === "1"
        ? "Sign in to like this post"
        : "Sign in to dislike this post";
  });
}

document.addEventListener("click", async function (event) {
  const button = event.target.closest(".post-card .reaction-button[data-post-id]");
  if (!button || button.disabled) return;

  const postCard = button.closest(".post-card");
  const actions = button.closest(".post-card-actions");
  const status = postCard.querySelector(".reaction-status");
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
    const likeButton = actions.querySelector('.reaction-button[data-reaction="1"]');
    const dislikeButton = actions.querySelector('.reaction-button[data-reaction="-1"]');
    likeButton.querySelector(".reaction-count").textContent = result.like_count;
    dislikeButton.querySelector(".reaction-count").textContent = result.dislike_count;

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
    actions.querySelectorAll(".reaction-button").forEach((reactionButton) => {
      reactionButton.disabled = false;
    });
  }
});

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
