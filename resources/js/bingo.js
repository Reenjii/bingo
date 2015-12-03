// Encrypt plaintext using key
function encrypt(key, plaintext) {
	return sjcl.encrypt(key, plaintext)
}

// Decrypt plaintext using key
function decrypt(key, cipher) {
	return sjcl.decrypt(key, cipher)
}

// Display or hide an element
function display(selector, show) {
	if ((typeof show === 'undefined') ? true : show) {
		$(selector).show();
	} else {
		$(selector).hide();
	}
}

// Display or hide paste form
function displayForm(show) {
	display('#form', show);
}

// Display or hide paste
function displayPaste(show) {
	display('#paste', show);
}

// Display or hide comments
function displayDiscussion(show) {
	display('#discussion', show);
}

// Display raw paste data
function raw() {
	var data = $('#meta-plain').val();
	var newDoc = document.open('text/html', 'replace');
	newDoc.write('<pre>' + data + '</pre>');
	newDoc.close();
}

// Clone displayed paste
function clone() {
	displayPaste(false);
	displayDiscussion(false);
	displayForm(true);

	// Erase the id and the key in url
	window.history.replaceState(document.title, document.title, baseURL());

	// Erase the id and the key in url
	//history.replaceState(document.title, document.title, scriptLocation());
	
	// write plaintext in paste form
	$('#form textarea').text($('#meta-plain').val());
}

// Display form for a new paste
function newPaste() {
	// Hide paste and display form
	displayPaste(false);
	displayDiscussion(false);
	displayForm(true);
	
	// Clear textarea
	$('#form textarea').text('');

	// Clean
	window.history.replaceState(document.title, document.title, baseURL());

	// Erase the id and the key in url
	//history.replaceState(document.title, document.title, scriptLocation());
}

/**
* @return the current script location (without search or hash part of the URL).
* eg. http://server.com/zero/?aaaa#bbbb --> http://server.com/zero/
*/
function baseURL() {
	return window.location.origin + "/";
	//var scriptLocation = window.location.href.substring(0,window.location.href.length - window.location.search.length - window.location.hash.length);
	//var hashIndex = scriptLocation.indexOf("#");
	//if (hashIndex !== -1) {
	//	scriptLocation = scriptLocation.substring(0, hashIndex)
	//}
	//return scriptLocation
}

// Get the hash part of the URL
function getHash() {
	var hashIndex = window.location.href.indexOf("#");
	if (hashIndex >= 0) {
		return window.location.href.substring(hashIndex + 1);
	} else {
		return "";
	}
}

// Send a new paste
function send() {
	// Get plaintext
	var plaintext = $('#form textarea').val();
	if (plaintext.length === 0) {
		return;
	}
	
	// Generate random key
	var randomkey = sjcl.codec.base64.fromBits(sjcl.random.randomWords(8, 0), 0);
	
	// Build data to send
	var data = {
		data: encrypt(randomkey, plaintext),
		expire: parseInt($('#form select[name=expire]').val()),
		burn: $('#form input[name=burn]').prop('checked'),
		discussion: $('#form input[name=discussion]').prop('checked'),
		highlight: $('#form input[name=highlight]').prop('checked')
	};

	// Send paste
	$.ajax({
		url: baseURL(),
		method: "POST",
		data: JSON.stringify(data),
		contentType: "application/json; charset=utf-8",
		accept: "application/json",
		error: function(jqXHR, textStatus, errorThrown) {
			console.log(jqXHR);
			if (textStatus === "error") {
				// The server replied with an HTTP error code
				displayDanger(jqXHR.responseJSON.error || "Oops, an error occurred.");
			} else {
				// An error occurred
				displayDanger("Oops, an error occurred.");
			}
		},
		success: function(response, textStatus, jqXHR) {
			// Build paste & delete URLs
			pasteUrl = baseURL() + response.id + "#" + randomkey;
			deleteUrl = baseURL() + "delete/" + response.id + "/" + response.delete;
			fillPasteUrl(pasteUrl, deleteUrl);
			
			paste = {
				id: response.id,
				data: data.data,
				plaintext: plaintext,
				postdate: response.postdate,
				expire: response.expire,
				burn: data.burn,
				discussion: data.discussion,
				highlight: data.highlight,
			};
			
			// Display paste
			fillPaste(paste);
			displayPaste(true);
			displayForm(false);
			
			// Update URL
			window.history.replaceState(document.title, document.title, pasteUrl);
		},
	});
}

// Send a new comment
function comment(parentid) {
	// Get plaintext
	var plaintext = $('#reply textarea').val();
	if (plaintext.length === 0) {
		return;
	}
	
	// Get paste id
	if (!paste) {
		return;
	}
	
	// Get author
	var author = $('#reply input[name=author]').val();
	
	// Get random key from URL
	var randomkey = getHash();
	
	// Build data to send
	var data = {
		data: encrypt(randomkey, plaintext),
		author: (author.length === 0) ? '' : encrypt(randomkey, author),
		highlight: $('#reply input[name=highlight]').prop('checked'),
		comment: true,
		parent: parentid,
		paste: paste.id,
	};

	// Send paste
	$.ajax({
		url: baseURL(),
		method: "POST",
		data: JSON.stringify(data),
		contentType: "application/json; charset=utf-8",
		accept: "application/json",
		error: function(jqXHR, textStatus, errorThrown) {
			console.log(jqXHR);
			if (textStatus === "error") {
				// The server replied with an HTTP error code
				displayDanger(jqXHR.responseJSON.error || "Oops, an error occurred.");
			} else {
				// An error occurred
				displayDanger("Oops, an error occurred.");
			}
		},
		success: function(response) {
			// Hide reply form
			$('#reply').remove();
			
			// Append comment
			appendComment({
				id: response.id,
				data: data.data,
				author: data.author,
				parent: parentid,
				highlight: data.highlight,
				postdate: response.postdate,
				avatar: response.avatar,
			});
		},
	});
}

function formatDate(date) {
	var month = ("0" + (date.getMonth() + 1)).slice(-2);
	var day = ("0" + date.getDate()).slice(-2);
	var h = ("0" + date.getHours()).slice(-2);
	var m = ("0" + date.getMinutes()).slice(-2);
	var s = ("0" + date.getSeconds()).slice(-2);
	var d = date.getFullYear() + '-' + month + '-' + day;
	var t = h + ":" + m + ":" + s;
	return d + " " + t; 
}

// Fill paste data
function fillPaste(paste) {
	// Fill paste data
	var div = $('#paste #data');
	if (paste.highlight) {
		div.html('<pre><code>' + he.escape(paste.plaintext) + '</code></pre>');
		div.each(function(i, block) {
			hljs.highlightBlock(block);
		});
		div.css('padding', '0');
	} else {
		div.html(he.escape(paste.plaintext).replace(/\n/ig,"<br>"));
		//div.html(he.escape(paste.plaintext));
	}
	
	// Fill paste expiration date
	if (paste.expire) {
		$('#paste-expire').html(formatDate(new Date(paste.expire)));
	} else {
		$('#paste-expire').hide();
	}
	
	// Fill paste creation date
	if (paste.postdate) {
		$('#paste-postdate').html(formatDate(new Date(paste.postdate)));
	} else {
		$('#paste-postdate').hide();
	}
	
	// Fill metadata plaintext
	$('#meta-plain').text(paste.plaintext);
	
	// Configure comment button
	$('#paste-comment').click(function() { displayCommentForm($('#paste-comment'), ''); });
	
	// Fill paste comments
	if (paste.comments) {
		paste.comments.map(appendComment);
	}
	
	displayDiscussion(paste.discussion);
}

function appendComment(comment) {
	// Decipher comment data
	var plain = decrypt(getHash(), comment.data);
	var anonymous = comment.author.length === 0;
	var plainauthor = anonymous ? '(Anonymous)' : decrypt(getHash(), comment.author);
	
	// Default parent block is root comments container
	var parentBlock = $('#comments');
	
	// Retrieve & clone comment template
	var div = $('#template-comment').children().first().clone();

	// Set comment div id
	div.attr('id', 'comment_' + comment.id);

	// Fill comment meta
	var meta = div.find('.comment-meta');
	meta.find('.comment-meta-author').html('<img src="data:image/png;base64,' + comment.avatar + '"> ' + plainauthor);
	if (anonymous) {
		meta.find('.comment-meta-author').css('color','red');
	}
	meta.find('.comment-meta-postdate').html(formatDate(new Date(comment.postdate)));

	// Fill comment data
	var datadiv = div.find('.comment-data');
	if (comment.highlight) {
		datadiv.html('<pre><code>' + he.escape(plain) + '</code></pre>');
		datadiv.each(function(i, block) {
			hljs.highlightBlock(block);
		});
		datadiv.css('padding', '0');
	} else {
		datadiv.html(he.escape(plain).replace(/\n/ig,"<br>"));
		//datadiv.html(he.escape(plain));
	}

	// Bind reply button click
	var replydiv = div.find('.comment-reply');
	div.find('button').click(function() { displayCommentForm(replydiv, comment.id); });
	
	// Find parent block if this comment is a reply
	var parentBlockId = '#comment_' + comment.parent;
	if ($(parentBlockId).length) {
		parentBlock = $(parentBlockId)
	}
	
	// Append block
	parentBlock.append(div);
}

function displayCommentForm(e, parentid) {
	// Remove any other reply form
	$('#reply').remove();
	
	// Retrieve & clone reply form template
	var div = $('#template-reply').children().first().clone();

	// Set div id
	div.attr('id', 'reply');
	
	// Bind button click
	div.find('button').click(function () { comment(parentid); });
	
	// Append comment form
	e.after(div);
}

function fillPasteUrl(pasteUrl, deleteUrl) {
	displaySuccess('Your paste url is <a href="' + pasteUrl + '" class="alert-link">' + pasteUrl + '</a>');
	displayInfo('This paste can be deleted using <a href="' + deleteUrl + '" class="alert-link">' + deleteUrl + '</a>');
}

function displayAlert(message, type) {
	// Retrieve alert template
	var div = $('#template-alert').children().first().clone();
	// Add class
	div.addClass('alert-' + (type || 'info'));
	// Add content
	div.find('button').after(message);
	// Display in alert container
	$('#alerts').append(div);
}

function displaySuccess(message) {
	displayAlert(message, 'success');
}

function displayInfo(message) {
	displayAlert(message, 'info');
}

function displayWarning(message) {
	displayAlert(message, 'warning');
}

function displayDanger(message) {
	displayAlert(message, 'danger');
}

// globals
var paste;

// on load
$(function() {

	// Assume there is no pase and fisplay form
	displayPaste(false);
	displayDiscussion(false);
	displayForm(true);

	// Display paste if any
	var pasteJSON = $('#meta #meta-paste').val();
	if (pasteJSON.length > 0) {
		paste = $.parseJSON(pasteJSON);

		// Decrypt cipher
		paste.plaintext = decrypt(getHash(), paste.data);
		
		// Fill paste
		fillPaste(paste);
		
		// Display paste
		displayPaste(true);
		displayForm(false);
	}
});
