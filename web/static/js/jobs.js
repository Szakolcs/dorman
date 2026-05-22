function openJobCreateModal() {
	const dialog = document.getElementById("job-create-dialog");
	if (dialog) {
		dialog.showModal();
	}
}

function closeJobCreateModal() {
	const dialog = document.getElementById("job-create-dialog");
	if (dialog) {
		dialog.close();
	}
}

document.addEventListener("DOMContentLoaded", () => {
	document.body.addEventListener("htmx:afterRequest", (event) => {
		const xhr = event.detail.xhr;
		if (!event.detail.successful || !xhr) {
			return;
		}
		if (xhr.getResponseHeader("HX-Redirect")) {
			closeJobCreateModal();
		}
	});
});
