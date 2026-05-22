function openInventoryCreateModal() {
	const dialog = document.getElementById("inventory-create-dialog");
	if (dialog) {
		dialog.showModal();
	}
}

function closeInventoryCreateModal() {
	const dialog = document.getElementById("inventory-create-dialog");
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
			closeInventoryCreateModal();
		}
	});
});
