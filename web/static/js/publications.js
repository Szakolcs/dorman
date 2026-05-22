document.addEventListener("DOMContentLoaded", () => {
	document.body.addEventListener("htmx:afterRequest", (event) => {
		const xhr = event.detail.xhr;
		if (!event.detail.successful || !xhr) {
			return;
		}
		const redirect = xhr.getResponseHeader("HX-Redirect");
		if (redirect) {
			window.location.assign(redirect);
		}
	});
});
