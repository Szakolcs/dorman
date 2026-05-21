.PHONY: css css-watch

# Tailwind CSS → web/static/css/app.css (requires: npm install)
css:
	npm run build:css

css-watch:
	npm run watch:css
