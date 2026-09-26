/**
 * Prevent TDesign icons from loading from its CDN. The icon package inserts
 * remote script and stylesheet tags on mount. Matching inert tags satisfy
 * its duplicate checks, while the sprite is loaded locally in index.html.
 */

const SVG_SCRIPT_CLASS = "t-svg-js-stylesheet--unique-class";
const ICONFONT_LINK_CLASS = "t-iconfont-stylesheet--unique-class";

// Match the URLs embedded in supported TDesign icon versions.
const BLOCKED_ICON_VERSIONS = ["0.4.0", "0.4.1", "0.4.2", "0.4.3", "0.4.4"];

const BLOCKED_SCRIPT_URLS = BLOCKED_ICON_VERSIONS.map(
  (version) => `https://tdesign.gtimg.com/icon/${version}/fonts/index.js`,
);

const BLOCKED_LINK_URLS = BLOCKED_ICON_VERSIONS.map(
  (version) => `https://tdesign.gtimg.com/icon/${version}/fonts/index.css`,
);

let installed = false;

export function installTDesignIconOfflineGuard(): void {
  if (installed || typeof document === "undefined") return;
  installed = true;

  const body = document.body;
  if (!body) {
    document.addEventListener(
      "DOMContentLoaded",
      () => installTDesignIconOfflineGuard(),
      { once: true },
    );
    installed = false;
    return;
  }

  BLOCKED_SCRIPT_URLS.forEach((src) => {
    const exists = document.querySelector(
      `script.${SVG_SCRIPT_CLASS}[src="${src}"]`,
    );
    if (exists) return;
    const stub = document.createElement("script");
    stub.setAttribute("class", SVG_SCRIPT_CLASS);
    stub.setAttribute("src", src);
    // An unknown MIME type prevents the browser from fetching this script.
    stub.setAttribute("type", "text/no-load");
    stub.setAttribute("data-knowledge-hub-blocked-cdn", "tdesign-icons");
    body.appendChild(stub);
  });

  BLOCKED_LINK_URLS.forEach((href) => {
    const exists = document.querySelector(
      `link.${ICONFONT_LINK_CLASS}[href="${href}"]`,
    );
    if (exists) return;
    const stub = document.createElement("link");
    stub.setAttribute("class", ICONFONT_LINK_CLASS);
    stub.setAttribute("href", href);
    // An unknown rel prevents stylesheet fetching.
    stub.setAttribute("rel", "preload-blocked");
    stub.setAttribute("data-knowledge-hub-blocked-cdn", "tdesign-icons");
    document.head.appendChild(stub);
  });
}
