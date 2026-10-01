"use strict";

(() => {
    const input = document.getElementById("image-file");
    const preview = document.getElementById("image-preview");
    if (!input || !preview) return;

    const original = Array.from(preview.childNodes, node => node.cloneNode(true));
    let previewURL;

    function releasePreview() {
        if (previewURL) URL.revokeObjectURL(previewURL);
        previewURL = undefined;
    }

    function showPreview() {
        releasePreview();
        const file = input.files[0];
        if (!file) {
            preview.replaceChildren(...original.map(node => node.cloneNode(true)));
            return;
        }

        const name = document.createElement("p");
        name.textContent = file.name;
        if (file.size > 10 * 1024 * 1024) {
            name.textContent = "Image exceeds 10 MiB. Choose a smaller image.";
            preview.replaceChildren(name);
            return;
        }

        const img = document.createElement("img");
        img.alt = "Selected image preview";
        const url = URL.createObjectURL(file);
        previewURL = url;
        img.addEventListener("error", () => {
            if (previewURL !== url) return;
            releasePreview();
            name.textContent = "Image could not be previewed. Choose a valid PNG, JPEG, or GIF.";
            preview.replaceChildren(name);
        });
        img.src = url;
        preview.replaceChildren(img, name);
    }

    input.addEventListener("change", showPreview);

    window.addEventListener("pagehide", releasePreview);
    // History restoration keeps the file selection but its old URL was revoked.
    window.addEventListener("pageshow", event => {
        if (event.persisted) showPreview();
    });
})();
