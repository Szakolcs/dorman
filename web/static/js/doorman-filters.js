function isAccessFilterForm(form) {
	return form && form.classList.contains("dm-access-filter-bar");
}

function markFilterFieldActive(input) {
	if (!input || !input.name) {
		return;
	}
	input.dataset.filterActive = "true";
	if (input.name === "from" || input.name === "from_hour") {
		var pair = input.form.querySelector('[name="' + (input.name === "from" ? "from_hour" : "from") + '"]');
		if (pair) {
			pair.dataset.filterActive = "true";
		}
	}
	if (input.name === "to" || input.name === "to_hour") {
		var pair = input.form.querySelector('[name="' + (input.name === "to" ? "to_hour" : "to") + '"]');
		if (pair) {
			pair.dataset.filterActive = "true";
		}
	}
}

document.addEventListener("change", function (event) {
	var input = event.target;
	if (!(input instanceof HTMLInputElement) || !input.form || !isAccessFilterForm(input.form)) {
		return;
	}
	if (input.type !== "date" && input.type !== "time") {
		return;
	}
	markFilterFieldActive(input);
});

document.addEventListener("click", function (event) {
	var button = event.target.closest("[data-access-status]");
	if (!button) {
		return;
	}

	var group = button.closest("[data-access-status-filters]");
	if (!group) {
		return;
	}

	var input = group.querySelector('input[name="status"]');
	if (!input) {
		return;
	}

	var status = button.getAttribute("data-access-status");
	if (input.value !== status) {
		group.querySelectorAll("[data-access-status]").forEach(function (item) {
			item.classList.toggle("dm-tenant-kind-filter-active", item === button);
		});
		input.value = status;
	}

	var form = group.closest("form");
	if (!form || !window.htmx) {
		return;
	}

	var trigger = form.getAttribute("data-filter-trigger");
	if (trigger) {
		window.htmx.trigger(form, trigger);
	}
});

document.body.addEventListener("htmx:configRequest", function (event) {
	var elt = event.detail.elt;
	var form = elt && elt.tagName === "FORM" ? elt : elt && elt.closest ? elt.closest("form") : null;
	if (!form || !isAccessFilterForm(form)) {
		return;
	}

	form.querySelectorAll('input[type="date"], input[type="time"]').forEach(function (input) {
		if (input.dataset.filterActive !== "true") {
			delete event.detail.parameters[input.name];
		}
	});
});
