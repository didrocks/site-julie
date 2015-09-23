/*
Copyright (c) 2015 The Polymer Project Authors. All rights reserved.
This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
Code distributed by Google as part of the polymer project is also
subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
*/

(function(document) {
  'use strict';

  // Grab a reference to our auto-binding template
  // and give it some initial binding values
  // Learn more about auto-binding templates at http://goo.gl/Dx1u2g
  let app = document.querySelector('#app');

  /* TODO: send a pull request to polymer starter kit for this
    save header y position between page to restore it at the same place.
    This enables to only condense and not condense header with a smooth transition without restoring to full
    size first. */
  let headerY = 0;

  app.displayInstalledToast = () => {
    // Check to make sure caching is actually enabled—it won't be in the dev environment.
    if (!document.querySelector('platinum-sw-cache').disabled) {
      document.querySelector('#caching-complete').show();
    }
  };

  // Listen for template bound event to know when bindings
  // have resolved and content has been stamped to the page
  app.addEventListener('dom-change', () => {
    console.log('Our app is ready to rock!');
  });

  // See https://github.com/Polymer/polymer/issues/1381
  window.addEventListener('WebComponentsReady', () => {
    // imports are loaded and elements have been registered
  });

  // Main area's paper-scroll-header-panel custom condensing transformation of
  // the appName in the middle-container and the bottom title in the bottom-container.
  // The appName is moved to top and shrunk on condensing. The bottom sub title
  // is shrunk to nothing on condensing.
  addEventListener('paper-header-transform', e => {
    let appName = document.querySelector('#mainToolbar .app-name');
    let middleContainer = document.querySelector('#mainToolbar .middle-container');
    let bottomTitle = document.querySelector('#mainToolbar .bottom-title');
    let detail = e.detail;
    let heightDiff = detail.height - detail.condensedHeight;
    let yRatio = Math.min(1, detail.y / heightDiff);
    // appName max size when condensed. The smaller the number the smaller the condensed size.
    let maxMiddleScale = 0.75;
    let scaleMiddle = Math.max(maxMiddleScale,
                               (heightDiff - detail.y) / (heightDiff / (1 - maxMiddleScale))  + maxMiddleScale);
    let scaleBottom = 1 - yRatio;

    // Save last header Y position
    headerY = e.detail.y;

    // Move/translate middleContainer
    Polymer.Base.transform(`translate3d(0, ${yRatio * 100}%, 0)`, middleContainer);

    // Scale bottomContainer and bottom sub title to nothing and back
    Polymer.Base.transform(`scale(${scaleBottom}) translateZ(0)`, bottomTitle);

    // Scale middleContainer appName
    Polymer.Base.transform(`scale(${scaleMiddle}) translateZ(0)`, appName);
  });

  // Scroll page to top and expand header
  app.scrollPageToTop = condenseHeader => {
    let headerPanel = document.querySelector('paper-scroll-header-panel[main]');
    // this allows to scroll the content to top, even if the final result is condensed
    headerPanel.scroll(headerY, false);
    if (condenseHeader) {
      headerPanel.condense(true);
    } else {
      headerPanel.scroll(0, true);
    }
  };

  document.addEventListener('WebComponentsReady', () => {
    // initial load for home page to trigger loading animation
    document.querySelector('home-page').show();
  });

})(document);
