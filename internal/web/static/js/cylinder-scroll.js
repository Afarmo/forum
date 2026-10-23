const carousel = document.getElementById('carousel');
const cells = carousel.querySelectorAll('.carousel-cell');
const cellCount = cells.length;
const theta = 360 / cellCount;
const cellHeight = 56;

// Optimal radius calculation
const radius = Math.round(
  (cellHeight / 2) /
  Math.tan(Math.PI / cellCount)
);

let currentIndex = 0;

// Position each cell in 3D space
cells.forEach((cell, i) => {
  const angle = theta * i * -1;
  cell.style.transform =
    `rotateX(${angle}deg) translateZ(${radius}px)`;
});

// Carousel rotation
function rotateCarousel(direction) {
  const nextIndex = currentIndex + direction;
  if (nextIndex < 0 || nextIndex >= cellCount) return false;

  currentIndex = nextIndex;
  const angle = theta * currentIndex;
  carousel.style.transform =
    `rotateX(${angle}deg)`;
  return true;
}

function isCarouselAtEnd() {
  return currentIndex === cellCount - 1;
}