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
	if (input.value === status) {
		return;
	}

	group.querySelectorAll("[data-access-status]").forEach(function (item) {
		item.classList.toggle("dm-tenant-kind-filter-active", item === button);
	});
	input.value = status;

	var form = group.closest("form");
	if (!form || !window.htmx) {
		return;
	}

	var trigger = form.getAttribute("data-filter-trigger");
	if (trigger) {
		window.htmx.trigger(form, trigger);
	}
});
