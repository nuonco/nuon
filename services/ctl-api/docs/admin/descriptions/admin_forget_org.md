Forget an org by soft-deleting it. This sets `deleted_at` on the org (and its roles) to hide the record from normal queries — the underlying row and related data are not removed from the database.

Event loops are designed to shut themselves down when their tracking object is deleted.
