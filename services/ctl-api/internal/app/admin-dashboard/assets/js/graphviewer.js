// Graph Viewer Component - Renders Graphviz DOT strings to SVG
// Uses @hpcc-js/wasm for client-side rendering

(async function() {
    'use strict';

    let graphvizInstance = null;

    // Load @hpcc-js/wasm from CDN
    async function loadGraphviz() {
        if (graphvizInstance) return graphvizInstance;

        try {
            // Dynamically import @hpcc-js/wasm from CDN
            const module = await import('https://cdn.jsdelivr.net/npm/@hpcc-js/wasm@2.13.0/dist/index.js');
            graphvizInstance = await module.Graphviz.load();
            return graphvizInstance;
        } catch (error) {
            console.error('GraphViewer: Failed to load Graphviz', error);
            throw error;
        }
    }

    // Initialize all graph viewers on the page
    async function initGraphViewers() {
        const viewers = document.querySelectorAll('[data-graph-viewer]');
        if (viewers.length === 0) return;

        try {
            const graphviz = await loadGraphviz();

            viewers.forEach(viewer => {
                const dotString = viewer.dataset.dotString;
                const viewerId = viewer.dataset.viewerId;

                if (!dotString) {
                    console.error('GraphViewer: No DOT string provided');
                    return;
                }

                setupViewer(viewer, dotString, graphviz, viewerId);
            });
        } catch (error) {
            console.error('GraphViewer: Initialization failed', error);
        }
    }

    // Set up a single viewer instance
    function setupViewer(container, dotString, graphviz, viewerId) {
        const svgContainer = container.querySelector('[data-graph-svg]');
        const dotContainer = container.querySelector('[data-graph-dot]');
        const toggleBtn = container.querySelector('[data-graph-toggle]');
        const toggleText = toggleBtn?.querySelector('[data-graph-toggle-text]');
        const errorContainer = container.querySelector('[data-graph-error]');

        let currentView = 'svg'; // Start with SVG view

        // Render DOT to SVG
        function renderSVG() {
            try {
                errorContainer.classList.add('hidden');
                const svg = graphviz.dot(dotString);
                svgContainer.innerHTML = svg;

                // Style the SVG for better display
                const svgElement = svgContainer.querySelector('svg');
                if (svgElement) {
                    svgElement.style.width = '100%';
                    svgElement.style.height = 'auto';
                    svgElement.style.maxHeight = '600px';

                    // Apply custom styling for dark theme
                    styleSVGElements(svgElement);
                }
            } catch (error) {
                console.error('GraphViewer: Render failed', error);
                errorContainer.textContent = `Rendering failed: ${error.message}`;
                errorContainer.classList.remove('hidden');
            }
        }

        // Apply custom styling to SVG elements for better visual appearance
        function styleSVGElements(svgElement) {
            // Style nodes (blue rectangles with white text)
            const nodes = svgElement.querySelectorAll('.node');
            nodes.forEach(node => {
                // Get the polygon/path elements (node shapes)
                const shapes = node.querySelectorAll('polygon, path, rect');
                shapes.forEach(shape => {
                    const color = node.querySelector('polygon, path')?.getAttribute('stroke') || 'blue';

                    if (color === 'blue') {
                        // Configured components - blue
                        shape.setAttribute('fill', '#3b82f6');
                        shape.setAttribute('stroke', '#3b82f6');
                    } else {
                        // Dependencies - lighter blue or gray
                        shape.setAttribute('fill', '#6366f1');
                        shape.setAttribute('stroke', '#6366f1');
                    }
                    shape.setAttribute('stroke-width', '1');
                    // Remove any rounded corners
                    shape.removeAttribute('rx');
                    shape.removeAttribute('ry');
                    // Remove any shadows or filters
                    shape.removeAttribute('filter');
                    shape.style.filter = 'none';
                    shape.style.boxShadow = 'none';
                });

                // Style text (white color, monospace font)
                const texts = node.querySelectorAll('text');
                texts.forEach(text => {
                    text.setAttribute('fill', '#ffffff');
                    text.setAttribute('font-family', 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace');
                    text.setAttribute('font-size', '11');
                    text.setAttribute('font-weight', '400');
                });
            });

            // Style edges (red/pink lines)
            const edges = svgElement.querySelectorAll('.edge');
            edges.forEach(edge => {
                const paths = edge.querySelectorAll('path');
                paths.forEach(path => {
                    path.setAttribute('stroke', '#ef4444');
                    path.setAttribute('stroke-width', '2');
                    path.setAttribute('fill', 'none');
                });

                // Style arrowheads
                const polygons = edge.querySelectorAll('polygon');
                polygons.forEach(polygon => {
                    polygon.setAttribute('fill', '#ef4444');
                    polygon.setAttribute('stroke', '#ef4444');
                });
            });

            // Remove default graph background
            const graphPolygon = svgElement.querySelector('g.graph > polygon');
            if (graphPolygon) {
                graphPolygon.setAttribute('fill', 'transparent');
            }
        }

        // Toggle between SVG and DOT views
        function toggleView() {
            if (currentView === 'svg') {
                // Switch to DOT view
                svgContainer.classList.add('hidden');
                dotContainer.classList.remove('hidden');
                toggleText.textContent = 'Show Graph';
                currentView = 'dot';
            } else {
                // Switch to SVG view
                dotContainer.classList.add('hidden');
                svgContainer.classList.remove('hidden');
                toggleText.textContent = 'Show Source';
                currentView = 'svg';
            }
        }

        // Set up toggle button
        if (toggleBtn) {
            toggleBtn.addEventListener('click', toggleView);
        }

        // Initial render
        renderSVG();
    }

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initGraphViewers);
    } else {
        initGraphViewers();
    }

    // Re-initialize on HTMX swaps (if using HTMX)
    document.body.addEventListener('htmx:afterSwap', initGraphViewers);
})();
