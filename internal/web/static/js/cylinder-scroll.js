// Category carousel: items sit at even spacing like a normal list, and each
// one tilts back and fades the further it is from the front item, so the list
// reads as a drum. "All Categories" is the front item on every page load.
//
// Desktop: a vertical drum with the front row at the top.
// Mobile: a horizontal strip with the front item on the left, so the sidebar
// stays short — swipe sideways to move through it.
//
// Scrolling or swiping over the sidebar moves the list one item at a time and
// never scrolls the page behind it.
const carousel = document.getElementById('carousel');

if (carousel) {
  const scene = carousel.closest('.carousel-scene');
  const cells = Array.from(carousel.querySelectorAll('.carousel-cell'));
  const mobile = window.matchMedia('(max-width: 768px)');

  const ROW_HEIGHT = 56;       // px between row centres on desktop
  const ITEM_GAP = 20;         // px between items on mobile (items are sized to their text)
  const TILT_PER_ITEM = 9;     // degrees each item past the front tilts back
  const MAX_TILT = 63;
  const FADE_PER_ITEM = 0.11;
  const MIN_OPACITY = 0.15;
  const WHEEL_STEP = 60;       // px of wheel travel per item
  const WHEEL_LINE = 20;       // px per "line" for mice that report lines (Firefox)
  const SWIPE_STEP = 40;       // px of finger travel per item
  // Speed limit: at most one item per this many ms, so items finish their
  // move-and-fade animation and stay readable during a fast flick.
  const MIN_STEP_INTERVAL = 150;
  // Stop while this many items are still visible, so the list never ends on
  // a single item.
  const KEEP_VISIBLE_AT_END = 3;

  const maxIndex = Math.max(0, cells.length - KEEP_VISIBLE_AT_END);
  let currentIndex = 0;
  let lastStepTime = 0;

  // Distance from the front item to item i, along the list's direction
  function distanceFromFront(i, horizontal, widths) {
    if (!horizontal) return (i - currentIndex) * ROW_HEIGHT;

    const from = Math.min(i, currentIndex);
    const to = Math.max(i, currentIndex);
    let distance = 0;
    for (let k = from; k < to; k++) distance += widths[k] + ITEM_GAP;
    return i < currentIndex ? -distance : distance;
  }

  function layout() {
    const horizontal = mobile.matches;
    const widths = horizontal ? cells.map((cell) => cell.offsetWidth) : null;

    cells.forEach((cell, i) => {
      const offset = i - currentIndex;
      const distance = distanceFromFront(i, horizontal, widths);
      const tilt = Math.min(Math.max(offset, 0) * TILT_PER_ITEM, MAX_TILT);

      cell.style.transform = horizontal
        ? `translateX(${distance}px) rotateY(${tilt}deg)`
        : `translateY(${distance}px) rotateX(${-tilt}deg)`;
      // Items already moved past slide out of view and fade out
      cell.style.opacity =
        offset < 0 ? 0 : Math.max(MIN_OPACITY, 1 - offset * FADE_PER_ITEM);
    });
  }

  function step(direction) {
    const now = performance.now();
    if (now - lastStepTime < MIN_STEP_INTERVAL) return;

    const next = Math.min(Math.max(currentIndex + direction, 0), maxIndex);
    if (next === currentIndex) return;

    currentIndex = next;
    lastStepTime = now;
    layout();
  }

  layout();
  mobile.addEventListener('change', layout);
  window.addEventListener('resize', layout);

  // Wheel: always swallow the event so the feed never scrolls while the
  // pointer is over the sidebar. Small trackpad deltas add up to one item;
  // anything beyond that is dropped rather than queued, so the list stops
  // when the scrolling stops.
  let wheelDistance = 0;

  scene.addEventListener('wheel', (event) => {
    event.preventDefault();

    let delta =
      Math.abs(event.deltaX) > Math.abs(event.deltaY) ? event.deltaX : event.deltaY;
    if (event.deltaMode === WheelEvent.DOM_DELTA_LINE) delta *= WHEEL_LINE;
    if (event.deltaMode === WheelEvent.DOM_DELTA_PAGE) delta *= carousel.clientHeight;

    if (Math.sign(delta) !== Math.sign(wheelDistance)) wheelDistance = 0;
    wheelDistance += delta;

    if (Math.abs(wheelDistance) >= WHEEL_STEP) {
      step(Math.sign(wheelDistance));
      wheelDistance = 0;
    }
  }, { passive: false });

  // Touch: drag the list with the finger — up on desktop, left on mobile
  // brings later items to the front. touch-action in the CSS leaves the other
  // direction to the browser.
  let touchPos = null;

  function touchPoint(event) {
    const touch = event.touches[0];
    return mobile.matches ? touch.clientX : touch.clientY;
  }

  scene.addEventListener('touchstart', (event) => {
    touchPos = touchPoint(event);
  }, { passive: true });

  scene.addEventListener('touchmove', (event) => {
    if (touchPos === null) return;
    const pos = touchPoint(event);
    const moved = touchPos - pos;

    if (Math.abs(moved) >= SWIPE_STEP) {
      step(Math.sign(moved));
      touchPos = pos;
    }
  }, { passive: true });

  scene.addEventListener('touchend', () => {
    touchPos = null;
  });
}
