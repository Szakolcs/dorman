document.addEventListener("DOMContentLoaded", () => {
	document.querySelectorAll(".dm-table-row-link[data-row-href]").forEach((row) => {
		row.addEventListener("click", () => {
			window.location.href = row.dataset.rowHref;
		});
	});
});
