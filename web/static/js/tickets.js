function openTicketCreateModal() {
	const dialog = document.getElementById("ticket-create-dialog");
	if (dialog) {
		dialog.showModal();
	}
}

function closeTicketCreateModal() {
	const dialog = document.getElementById("ticket-create-dialog");
	if (dialog) {
		dialog.close();
	}
}

function initTicketsPage() {
	document.body.addEventListener("htmx:afterRequest", (event) => {
		if (!event.detail.successful) {
			return;
		}
		const form = event.detail.elt;
		if (form && form.closest && form.closest("#ticket-create-dialog")) {
			closeTicketCreateModal();
		}
	});
}

if (document.readyState === "loading") {
	document.addEventListener("DOMContentLoaded", initTicketsPage);
} else {
	initTicketsPage();
}
