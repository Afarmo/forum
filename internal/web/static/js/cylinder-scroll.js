const carousel = document.getElementById('carousel');
const cells = carousel.querySelectorAll('.carousel-cell');
const cellCount = cells.length;
const theta = 360 / cellCount;
const cellWidth = 140;

// Optimal radius calculation
const radius = Math.round(
  (cellWidth / 2) /
  Math.tan(Math.PI / cellCount)
);

let currentIndex = 0;

// Position each cell in 3D space
cells.forEach((cell, i) => {
  const angle = theta * i;
  cell.style.transform =
    `rotateY(${angle}deg) translateZ(${radius}px)`;
});

// Carousel rotation
function rotateCarousel(direction) {
  currentIndex += direction;
  const angle = theta * currentIndex * -1;
  carousel.style.transform =
    `rotateY(${angle}deg)`;
}