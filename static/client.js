'use strict';

function attr(el, init_attrs) {
	if (init_attrs) {
		for (var k in init_attrs) {
			el.setAttribute(k, init_attrs[k]);
		}
	}
}

function el(parent, tagName, init_attrs, text) {
	var doc = parent ? parent.ownerDocument : document;
	var el = doc.createElement(tagName);
	if (typeof init_attrs === 'string') {
		init_attrs = {
			'class': init_attrs,
		};
	}
	attr(el, init_attrs);
	if ((text !== undefined) && (text !== null)) {
		el.appendChild(doc.createTextNode(text));
	}
	if (parent) {
		parent.appendChild(el);
	}
	return el;
}

document.addEventListener('DOMContentLoaded', () => {
	// Create checkbox for every physical button
	const alllButtons = JSON.parse(document.querySelector('body').getAttribute('data-buttonJSON'));
	const buttonsList = document.querySelector('#buttons-list');
	let container = null;
	let lastPrefix = null;
	for (const btn of alllButtons) {
		const prefix = btn.split(/\s+/)[0];
		if (!container || (prefix !== lastPrefix)) {
			container = el(buttonsList, 'div');
			lastPrefix = prefix;
		}

		const label = el(container, 'label');
		el(label, 'input', {type: 'checkbox', name: 'button', value: btn});
		label.appendChild(document.createTextNode(btn));
	}

	const buttonsForm = document.querySelector('#buttons-form');
	buttonsForm.addEventListener('submit', async e => {
		e.preventDefault();

		const fd = new FormData(buttonsForm);
		const buttons = Array.from(fd.entries()).filter(([key, _]) => key === 'button').map(([_, btn]) => btn);

		// Send fetch request
		const response = await fetch('press-buttons', {
			method: 'POST',
			body: JSON.stringify({buttons}),
		});
		if (response.status !== 200) alert(`HTTP errorr ${response.status}`);
	});
});