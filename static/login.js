const loginBtn = document.getElementById('login');
const signupBtn = document.getElementById('signup');
const errorElement = document.querySelector('.error-message');
const closeButton = document.getElementById('close-button');

loginBtn.addEventListener('click', (e) => {
  let parent = e.target.parentNode.parentNode;
  Array.from(e.target.parentNode.parentNode.classList).find((element) => {
    if (element !== "slide-up") {
      parent.classList.add('slide-up');
    } else {
      signupBtn.parentNode.classList.add('slide-up');
      parent.classList.remove('slide-up');
    }
  });
});

signupBtn.addEventListener('click', (e) => {
  let parent = e.target.parentNode;
  Array.from(e.target.parentNode.classList).find((element) => {
    if (element !== "slide-up") {
      parent.classList.add('slide-up');
    } else {
      loginBtn.parentNode.parentNode.classList.add('slide-up');
      parent.classList.remove('slide-up');
    }
  });
});

// Function to show error message
function showError(message) {
  errorElement.classList.add('show');
}

// Function to hide error message
function hideError() {
  errorElement.classList.remove('show');
}

// Add event listener to close button
closeButton.addEventListener('click', hideError);

// Check if error message is present
if (errorElement) {
  showError();
}
