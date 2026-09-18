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

// Builds a post card DOM node from the JSON a post API call returns
// (see models.Post: content, title, post_id, user_id, picture_content, created_at).
// textContent is used throughout (never innerHTML) so post content can't inject markup.
function createPostCard(post) {
    const article = document.createElement("article");
    article.className = "post-card";

    const title = document.createElement("h3");
    title.className = "post-title";
    title.textContent = post.title;
    article.appendChild(title);

    if (post.picture_content) {
        const img = document.createElement("img");
        img.className = "post-image";
        img.src = "/static/img/post/" + post.picture_content;
        img.alt = post.title;
        article.appendChild(img);
    }

    const content = document.createElement("p");
    content.className = "post-content";
    content.textContent = post.content;
    article.appendChild(content);

    const date = document.createElement("time");
    date.className = "post-date";
    date.textContent = new Date(post.created_at).toLocaleString();
    article.appendChild(date);

    return article;
}

// Write-post form: submit via fetch so the response (created post JSON) can be
// turned straight into a card and dropped into the feed, instead of navigating
// the browser to the raw JSON the handler returns.
document.addEventListener("DOMContentLoaded", function () {
    const writePostForm = document.querySelector("#writePost-modal form");
    const feed = document.getElementById("post-feed");
    if (!writePostForm || !feed) return;

    writePostForm.addEventListener("submit", async function (e) {
        e.preventDefault();

        const res = await fetch(writePostForm.action, {
            method: "POST",
            body: new FormData(writePostForm),
        });

        if (!res.ok) {
            alert(await res.text());
            return;
        }

        const post = await res.json();
        feed.prepend(createPostCard(post));

        writePostForm.reset();
        closeModal("writePost-modal");
    });
});
document.addEventListener("DOMContentLoaded", async function () {
    const feed = document.getElementById("post-feed");
    if (!feed) return;

    const res = await fetch("/posts");
    if (!res.ok) return;

    const posts = await res.json();
    (posts || []).forEach(post => feed.appendChild(createPostCard(post)));
});
