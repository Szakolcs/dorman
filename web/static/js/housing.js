function openBuildingCreateModal() {
	const dialog = document.getElementById("building-create-dialog");
	if (dialog) {
		dialog.showModal();
	}
}

function closeBuildingCreateModal() {
	const dialog = document.getElementById("building-create-dialog");
	if (dialog) {
		dialog.close();
	}
}

function openFlatCreateModal() {
	const dialog = document.getElementById("flat-create-dialog");
	if (dialog) {
		dialog.showModal();
	}
}

function closeFlatCreateModal() {
	const dialog = document.getElementById("flat-create-dialog");
	if (dialog) {
		dialog.close();
	}
}

function openSharedAreaCreateModal() {
	const dialog = document.getElementById("shared-area-create-dialog");
	if (dialog) {
		dialog.showModal();
	}
}

function closeSharedAreaCreateModal() {
	const dialog = document.getElementById("shared-area-create-dialog");
	if (dialog) {
		dialog.close();
	}
}

function openRoomCreateModal() {
	const dialog = document.getElementById("room-create-dialog");
	if (dialog) {
		dialog.showModal();
	}
}

function closeRoomCreateModal() {
	const dialog = document.getElementById("room-create-dialog");
	if (dialog) {
		dialog.close();
	}
}

function openFlatInventoryCreateModal() {
	const dialog = document.getElementById("flat-inventory-create-dialog");
	if (dialog) {
		dialog.showModal();
	}
}

function closeFlatInventoryCreateModal() {
	const dialog = document.getElementById("flat-inventory-create-dialog");
	if (dialog) {
		dialog.close();
	}
}

function openSharedAreaInventoryCreateModal() {
	const dialog = document.getElementById("shared-area-inventory-create-dialog");
	if (dialog) {
		dialog.showModal();
	}
}

function closeSharedAreaInventoryCreateModal() {
	const dialog = document.getElementById("shared-area-inventory-create-dialog");
	if (dialog) {
		dialog.close();
	}
}

function closeHousingModals() {
	closeBuildingCreateModal();
	closeFlatCreateModal();
	closeSharedAreaCreateModal();
	closeRoomCreateModal();
	closeFlatInventoryCreateModal();
	closeSharedAreaInventoryCreateModal();
}

document.addEventListener("DOMContentLoaded", () => {
	document.body.addEventListener("htmx:afterRequest", (event) => {
		const xhr = event.detail.xhr;
		if (!event.detail.successful || !xhr) {
			return;
		}
		if (xhr.getResponseHeader("HX-Redirect")) {
			closeHousingModals();
		}
	});
});
