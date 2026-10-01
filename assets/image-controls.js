"use strict";

(() => {
    const form = document.getElementById("image-form");
    if (!form) return;

    function showValue(slider) {
        const text = `${slider.value} ${slider.dataset.unit}`.trim();
        slider.nextElementSibling.value = text;
        slider.setAttribute("aria-valuetext", text);
    }
    form.querySelectorAll('input[type="range"]').forEach(slider => {
        showValue(slider);
        slider.addEventListener("input", () => showValue(slider));
    });

    const result = document.getElementById("ascii-result");
    if (!result) return;

    const exports = document.getElementById("image-export");
    const status = document.getElementById("update-status");
    const file = document.getElementById("image-file");
    const dimensions = document.getElementById("dimensions");
    document.getElementById("settings-help").textContent =
        "Sliders and filters update the ASCII preview automatically. Filter strengths apply when their checkbox is checked. Transparent frame adds clear PNG padding.";

    let timer;
    let request;
    let revision = 0;

    function cancelUpdate() {
        clearTimeout(timer);
        request?.abort();
        revision++;
    }

    function setPending(pending) {
        result.setAttribute("aria-busy", String(pending));
        exports.querySelectorAll("button").forEach(button => { button.disabled = pending; });
    }

    async function update(currentRevision) {
        const controller = new AbortController();
        request = controller;
        // Reuse the bounded, retained source rather than re-uploading the file.
        const data = new FormData(form);
        data.delete("image-file");
        try {
            const response = await fetch(form.action, {
                method: "POST", body: data, signal: controller.signal,
            });
            const page = new DOMParser().parseFromString(await response.text(), "text/html");
            if (currentRevision !== revision) return;
            if (!response.ok) {
                throw new Error(page.querySelector(".panel p")?.textContent || "Preview could not be updated.");
            }
            const nextResult = page.getElementById("ascii-result");
            const nextDimensions = page.getElementById("dimensions");
            if (!nextResult || !nextDimensions) throw new Error("Preview could not be updated.");
            result.value = nextResult.value;
            dimensions.textContent = nextDimensions.textContent;
            exports.elements.namedItem("art").value = nextResult.value;
            exports.elements.namedItem("frame").value = data.get("frame");
            status.textContent = "Preview updated.";
            setPending(false);
        } catch (error) {
            if (currentRevision !== revision || error.name === "AbortError") return;
            result.setAttribute("aria-busy", "false");
            status.textContent = `${error.message} Use Apply settings to retry.`;
        } finally {
            if (request === controller) request = undefined;
        }
    }

    function scheduleUpdate() {
        cancelUpdate();
        setPending(true);
        if (file.files.length) {
            result.setAttribute("aria-busy", "false");
            status.textContent = "Apply settings to convert the newly selected image.";
            return;
        }
        status.textContent = "Updating preview…";
        const currentRevision = revision;
        timer = setTimeout(() => update(currentRevision), 180);
    }

    form.addEventListener("input", scheduleUpdate);
    form.addEventListener("change", scheduleUpdate);
    form.addEventListener("submit", cancelUpdate);
    window.addEventListener("pagehide", cancelUpdate);
    // Restore a cancelled pending update when returning via browser history.
    window.addEventListener("pageshow", event => { if (event.persisted) scheduleUpdate(); });
})();
