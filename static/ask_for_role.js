document.addEventListener('DOMContentLoaded', async () => {
    const token = localStorage.getItem('token');
    const user = JSON.parse(localStorage.getItem('user'));

    if (!token || !user) {
        alert('You need to log in first.');
        window.location.href = '/static/loginPage';
        return;
    }

    if (user.role !== 'admin') {
        alert('Access denied. You are not an admin.');
        window.location.href = '/';
        return;
    }

});