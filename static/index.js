document.addEventListener('DOMContentLoaded', function() {
    const toggleButtons = document.querySelectorAll('.toggle-button');
    toggleButtons.forEach(function(button) {
        button.addEventListener('click', function() {
            const playlistItems = this.parentNode.parentNode.querySelectorAll('.playlist-item');
            playlistItems.forEach(function(item) {
                item.classList.toggle('hidden');
            });
        });
    });
});
