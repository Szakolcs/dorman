function openPublicationCreateModal() {
	const dialog = document.getElementById("publication-create-dialog");
	if (dialog) {
		dialog.showModal();
	}
}

function closePublicationCreateModal() {
	const dialog = document.getElementById("publication-create-dialog");
	if (dialog) {
		dialog.close();
	}
}

function setPublicationFormKind(kind) {
	document.querySelectorAll("[data-publication-form]").forEach((panel) => {
		panel.hidden = panel.dataset.publicationForm !== kind;
	});
}

document.addEventListener("DOMContentLoaded", () => {
	const select = document.getElementById("publication-kind-select");
	if (select) {
		setPublicationFormKind(select.value);
		select.addEventListener("change", () => setPublicationFormKind(select.value));
	}

	document.body.addEventListener("htmx:afterRequest", (event) => {
		const xhr = event.detail.xhr;
		if (!event.detail.successful || !xhr) {
			return;
		}
		if (xhr.getResponseHeader("HX-Redirect")) {
			closePublicationCreateModal();
		}
	});
});
