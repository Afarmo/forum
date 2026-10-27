// Category list scroll hint. The list itself is scrolled by the browser (see
// scrollsidebar.css); this only nudges it once per session so users can see
// it scrolls: down and back on desktop, sideways and back on mobile.
const carousel = document.getElementById('carousel');

if (carousel) {
  const mobile = window.matchMedia('(max-width: 1100px)');

  const HINT_DISTANCE = 24;    // px the list moves before coming back
  const HINT_DELAY = 800;      // ms after load before the nudge
  const HINT_HOLD = 350;       // ms for each half of the nudge

  const hintTimers = [];

  // Any scroll, touch or click on the list stops the hint and gives the list
  // back to the user, with snapping turned on again.
  function cancelHint() {
    hintTimers.forEach(clearTimeout);
    hintTimers.length = 0;
    carousel.style.scrollSnapType = '';
  }

  function alreadyHinted() {
    try {
      if (sessionStorage.getItem('carouselHinted')) return true;
      sessionStorage.setItem('carouselHinted', '1');
    } catch (e) { /* storage blocked: just show the hint */ }
    return false;
  }

  function nudge(distance) {
    const by = mobile.matches ? { left: distance } : { top: distance };
    carousel.scrollBy({ ...by, behavior: 'smooth' });
  }

  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  const canScroll =
    carousel.scrollHeight > carousel.clientHeight ||
    carousel.scrollWidth > carousel.clientWidth;

  if (canScroll && !reducedMotion && !alreadyHinted()) {
    ['wheel', 'touchstart', 'pointerdown'].forEach((type) => {
      carousel.addEventListener(type, cancelHint, { once: true, passive: true });
    });

    // Snapping would pull the list straight back to the first item, so it's
    // switched off while the hint plays.
    hintTimers.push(setTimeout(() => {
      carousel.style.scrollSnapType = 'none';
      nudge(HINT_DISTANCE);
    }, HINT_DELAY));
    hintTimers.push(setTimeout(() => nudge(-HINT_DISTANCE), HINT_DELAY + HINT_HOLD));
    hintTimers.push(setTimeout(cancelHint, HINT_DELAY + 2 * HINT_HOLD));
  }
}
