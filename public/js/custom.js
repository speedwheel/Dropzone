Dropzone.options.myDropzone = {
    paramName: "file",
    addRemoveLinks : true,
    maxFiles:4,
    init: function() {
        thisDropzone = this;

        $.get('/uploads', function(data) {
            if (data == null) {
                return;
            }

            $.each(data, function(key, value) {
                var mockFile = {
                    name: value.name,
                    size: value.size,
                    status: Dropzone.ADDED,
                    accepted: true
                };
                thisDropzone.emit("addedfile", mockFile);
                thisDropzone.options.thumbnail.call(thisDropzone, mockFile, '/public/uploads/thumbnail_' + value.name);
                thisDropzone.emit("complete", mockFile);
                thisDropzone.files.push(mockFile);
            });
        });
    },
    renameFile: function (file) {
        file.newName = new Date().getTime() + '_' + file.name;
        return file.newName;
    },
    success: function(file, serverFileName) {
	       console.log(serverFileName);
	},
    removedfile: function(file) {
        if(file.accepted === false) {
            $(document).find(file.previewElement).remove();
            return;
        }
        var fileName = file.newName;
        if (fileName == null) {
            fileName = file.name;
        }
		$.ajax({
			type: 'DELETE',
			url: '/remove/'+fileName,
			dataType: 'json'
		}).done(
		function (data) {
			if(data === true) {
				$(document).find(file.previewElement).remove();
			}
		});
    }

};
