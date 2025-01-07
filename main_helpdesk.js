document.getElementById('supportForm').addEventListener('submit', async function (e) {
    e.preventDefault();
    
    const formData = new FormData(this);
    
    try {
        const response = await fetch('/send-support-ticket', {
            method: 'POST',
            body: formData
        });

        if (response.ok) {
            alert('Message sent successfully!');
        } else {
            alert('Failed to send message');
        }
    } catch (error) {
        console.error('Error sending message:', error);
        alert('Error sending message');
    }
});

window.onload = function() {
    const user = JSON.parse(localStorage.getItem('user'));
    
    if (user) {
        document.getElementById('login-btn').innerText = 'Logout';
        document.getElementById('profile-btn').innerText = 'Profile';

        document.getElementById('login-btn').onclick = function(e) {
            e.preventDefault(); 
            localStorage.removeItem('user');
            window.location.href = '/static/login_page.html';
        };

        document.getElementById('profile-btn').onclick = function(e) {
            e.preventDefault();
            window.location.href = '/static/profile.html';
        };
    } else {
        document.getElementById('login-btn').innerText = 'Login';
        document.getElementById('profile-btn').innerText = 'Sign-up';

        document.getElementById('login-btn').onclick = function() {
            window.location.href = '/static/login_page.html';
        };

        document.getElementById('profile-btn').onclick = function() {
            window.location.href = '/static/signup_page.html';
        };
    }
};

