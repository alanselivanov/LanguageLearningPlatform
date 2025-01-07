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