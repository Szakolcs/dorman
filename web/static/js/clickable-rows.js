const TABLE_ROW_INTERACTIVE =
	".dm-table-actions, button, a, form, input, select, textarea, label, [hx-post], [hx-delete], [hx-put], [hx-patch], [hx-get], .dm-mass-assignment-chip";

document.addEventListener("click", (event) => {
	const row = event.target.closest(".dm-table-row-link[data-row-href]");
	if (!row) {
		return;
	}
	if (event.defaultPrevented) {
		return;
	}
	if (event.target.closest(TABLE_ROW_INTERACTIVE)) {
		return;
	}

	window.location.href = row.dataset.rowHref;
});
