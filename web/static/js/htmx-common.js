// Shared HTMX response handling for all pages using components.Document.
function initHtmxCommon() {
	document.body.addEventListener("htmx:afterRequest", (event) => {
		const xhr = event.detail.xhr;
		if (!event.detail.successful || !xhr) {
			return;
		}

		const redirect = xhr.getResponseHeader("HX-Redirect");
		if (redirect) {
			window.location.assign(redirect);
			return;
		}

		if (xhr.getResponseHeader("HX-Refresh") === "true") {
			window.location.reload();
		}
	});
}

if (document.body) {
	initHtmxCommon();
} else {
	document.addEventListener("DOMContentLoaded", initHtmxCommon);
}
