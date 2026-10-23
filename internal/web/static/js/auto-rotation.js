let autoRotateInterval = null;
const AUTO_ROTATE_DELAY = 3000; // 3 secondes

// Start auto-rotation
function startAutoRotate() {
  stopAutoRotate();
  if (isCarouselAtEnd()) return;

  autoRotateInterval = setInterval(() => {
    if (!rotateCarousel(1) || isCarouselAtEnd()) {
      stopAutoRotate();
    }
  }, AUTO_ROTATE_DELAY);
}

// Stop auto-rotation
function stopAutoRotate() {
  if (autoRotateInterval) {
    clearInterval(autoRotateInterval);
    autoRotateInterval = null;
  }
}

// Pause on hover
const scene = document.querySelector('.carousel-scene');

scene.addEventListener('mouseenter', stopAutoRotate);
scene.addEventListener('mouseleave', startAutoRotate);

// Rotate the cylinder vertically when the pointer wheel is used over it.
scene.addEventListener('wheel', (event) => {
  if (event.deltaY === 0) return;

  const moved = rotateCarousel(event.deltaY > 0 ? 1 : -1);
  if (!moved) {
    if (isCarouselAtEnd()) stopAutoRotate();
    return;
  }

  event.preventDefault();
  stopAutoRotate();
  if (!isCarouselAtEnd()) setTimeout(startAutoRotate, 5000);
}, { passive: false });

// Pause on button click
document.querySelectorAll('.carousel-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    stopAutoRotate();
    // Resume after 5 seconds of inactivity
    setTimeout(startAutoRotate, 5000);
  });
});

// Start on load
startAutoRotate();

let touchStartX = 0;
let touchEndX = 0;
const SWIPE_THRESHOLD = 50; // pixels minimum

scene.addEventListener('touchstart', (e) => {
  touchStartX = e.changedTouches[0].screenX;
  stopAutoRotate();
}, { passive: true });

scene.addEventListener('touchend', (e) => {
  touchEndX = e.changedTouches[0].screenX;
  handleSwipe();
}, { passive: true });

function handleSwipe() {
  const diff = touchStartX - touchEndX;

  if (Math.abs(diff) > SWIPE_THRESHOLD) {
    if (diff > 0) {
      // Swipe left = next element
      rotateCarousel(1);
    } else {
      // Swipe right = previous element
      rotateCarousel(-1);
    }
  }

  // Resume auto-rotation after 5s
  setTimeout(startAutoRotate, 5000);
}