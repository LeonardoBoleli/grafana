package rest

import (
	"context"
	"errors"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog/v2"
)

type DualWriterMode1 struct {
	Legacy  LegacyStorage
	Storage Storage
	Log     klog.Logger
	*dualWriterMetrics
}

var mode = strconv.Itoa(int(Mode1))

// NewDualWriterMode1 returns a new DualWriter in mode 1.
// Mode 1 represents writing to and reading from LegacyStorage.
func NewDualWriterMode1(legacy LegacyStorage, storage Storage) *DualWriterMode1 {
	metrics := &dualWriterMetrics{}
	metrics.init()
	return &DualWriterMode1{Legacy: legacy, Storage: storage, Log: klog.NewKlogr().WithName("DualWriterMode1"), dualWriterMetrics: metrics}
}

// Create overrides the behavior of the generic DualWriter and writes only to LegacyStorage.
func (d *DualWriterMode1) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	log := d.Log.WithValues("kind", options.Kind)
	ctx = klog.NewContext(ctx, log)
	var method = "create"

	startLegacy := time.Now().UTC()
	res, err := d.Legacy.Create(ctx, obj, createValidation, options)
	if err != nil {
		log.Error(err, "unable to create object in legacy storage")
		d.recordLegacyDuration(true, mode, options.Kind, method, startLegacy)
		return res, err
	}
	d.recordLegacyDuration(false, mode, options.Kind, method, startLegacy)

	accessorCreated, err := meta.Accessor(res)
	if err != nil {
		log.Error(err, "unable to get accessor for created object")
	}

	accessorOld, err := meta.Accessor(obj)
	if err != nil {
		log.Error(err, "unable to get accessor for old object")
	}

	enrichObject(accessorOld, accessorCreated)

	go func() {
		startStorage := time.Now().UTC()
		ctx, _ := context.WithTimeoutCause(ctx, time.Second*10, errors.New("storage create timeout"))
		_, err := d.Storage.Create(ctx, obj, createValidation, options)
		defer d.recordStorageDuration(err != nil, mode, options.Kind, method, startStorage)
	}()

	return res, err
}

// Get overrides the behavior of the generic DualWriter and reads only from LegacyStorage.
func (d *DualWriterMode1) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	log := d.Log.WithValues("name", name, "resourceVersion", options.ResourceVersion, "kind", options.Kind)
	ctx = klog.NewContext(ctx, log)
	var method = "get"

	startLegacy := time.Now().UTC()
	res, err := d.Legacy.Get(ctx, name, options)
	if err != nil {
		log.Error(err, "unable to get object in legacy storage")
		d.recordLegacyDuration(true, mode, name, method, startLegacy)
		return res, err
	}
	d.recordLegacyDuration(false, mode, name, method, startLegacy)

	go func() {
		startStorage := time.Now().UTC()
		ctx, _ := context.WithTimeoutCause(ctx, time.Second*10, errors.New("storage get timeout"))
		_, err := d.Storage.Get(ctx, name, options)
		defer d.recordStorageDuration(err != nil, mode, name, method, startStorage)
	}()

	return res, err
}

// List overrides the behavior of the generic DualWriter and reads only from LegacyStorage.
func (d *DualWriterMode1) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	log := d.Log.WithValues("kind", options.Kind, "resourceVersion", options.ResourceVersion, "kind", options.Kind)
	ctx = klog.NewContext(ctx, log)
	var method = "list"

	startLegacy := time.Now().UTC()
	res, err := d.Legacy.List(ctx, options)
	if err != nil {
		log.Error(err, "unable to list object in legacy storage")
		d.recordLegacyDuration(true, mode, options.Kind, method, startLegacy)
		return res, err
	}
	d.recordLegacyDuration(false, mode, options.Kind, method, startLegacy)

	go func() {
		startStorage := time.Now().UTC()
		ctx, _ := context.WithTimeoutCause(ctx, time.Second*10, errors.New("storage list timeout"))
		_, err := d.Storage.List(ctx, options)
		defer d.recordStorageDuration(err != nil, mode, options.Kind, method, startStorage)
	}()

	return res, err
}

func (d *DualWriterMode1) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	log := d.Log.WithValues("name", name, "kind", options.Kind)
	ctx = klog.NewContext(ctx, d.Log)
	var method = "delete"

	startLegacy := time.Now().UTC()
	res, async, err := d.Legacy.Delete(ctx, name, deleteValidation, options)
	if err != nil {
		log.Error(err, "unable to delete object in legacy storage")
		d.recordLegacyDuration(true, mode, name, method, startLegacy)
		return res, async, err
	}
	d.recordLegacyDuration(false, mode, name, method, startLegacy)

	go func() {
		startStorage := time.Now().UTC()
		ctx, _ := context.WithTimeoutCause(ctx, time.Second*10, errors.New("storage delete timeout"))
		_, _, err := d.Storage.Delete(ctx, name, deleteValidation, options)
		defer d.recordStorageDuration(err != nil, mode, name, method, startStorage)
	}()

	return res, async, err
}

// DeleteCollection overrides the behavior of the generic DualWriter and deletes only from LegacyStorage.
func (d *DualWriterMode1) DeleteCollection(ctx context.Context, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions, listOptions *metainternalversion.ListOptions) (runtime.Object, error) {
	log := d.Log.WithValues("kind", options.Kind, "resourceVersion", listOptions.ResourceVersion)
	ctx = klog.NewContext(ctx, log)
	var method = "delete-collection"

	startLegacy := time.Now().UTC()
	res, err := d.Legacy.DeleteCollection(ctx, deleteValidation, options, listOptions)
	if err != nil {
		log.Error(err, "unable to delete collection in legacy storage")
		d.recordLegacyDuration(true, mode, options.Kind, method, startLegacy)
	}
	d.recordLegacyDuration(false, mode, options.Kind, method, startLegacy)

	go func() {
		startStorage := time.Now().UTC()
		ctx, _ := context.WithTimeoutCause(ctx, time.Second*10, errors.New("storage deletecollection timeout"))
		_, err := d.Storage.DeleteCollection(ctx, deleteValidation, options, listOptions)
		defer d.recordStorageDuration(err != nil, mode, options.Kind, method, startStorage)
	}()

	return res, err
}

func (d *DualWriterMode1) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	log := d.Log.WithValues("name", name, "kind", options.Kind)
	ctx = klog.NewContext(ctx, log)
	var method = "update"

	startLegacy := time.Now().UTC()
	res, async, errLegacy := d.Legacy.Update(ctx, name, objInfo, createValidation, updateValidation, forceAllowCreate, options)
	if errLegacy != nil {
		log.Error(errLegacy, "unable to update in legacy storage")
		d.recordLegacyDuration(true, mode, name, method, startLegacy)
	}
	d.recordLegacyDuration(false, mode, name, method, startLegacy)

	updated, err := objInfo.UpdatedObject(ctx, res)
	if err != nil {
		log.WithValues("object", updated).Error(err, "could not update or create object")
	}

	// get the object to be updated
	old, err := d.Storage.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		log.WithValues("object", old).Error(err, "could not get object to update")
	}

	// if the object is found, create a new updateWrapper with the object found
	if old != nil {
		objInfo = &updateWrapper{
			upstream: objInfo,
			updated:  old,
		}

		accessorOld, err := meta.Accessor(old)
		if err != nil {
			log.Error(err, "unable to get accessor for original updated object")
		}

		accessor, err := meta.Accessor(res)
		if err != nil {
			log.Error(err, "unable to get accessor for updated object")
		}

		accessor.SetResourceVersion(accessorOld.GetResourceVersion())
		accessor.SetUID(accessorOld.GetUID())

		enrichObject(accessorOld, accessor)
		objInfo = &updateWrapper{
			upstream: objInfo,
			updated:  res,
		}
	}

	go func() {
		startStorage := time.Now().UTC()
		ctx, _ := context.WithTimeoutCause(ctx, time.Second*10, errors.New("storage update timeout"))
		_, _, err := d.Storage.Update(ctx, name, objInfo, createValidation, updateValidation, forceAllowCreate, options)
		defer d.recordStorageDuration(err != nil, mode, name, method, startStorage)
	}()

	return res, async, errLegacy
}

func (d *DualWriterMode1) Destroy() {
	d.Storage.Destroy()
	d.Legacy.Destroy()
}

func (d *DualWriterMode1) GetSingularName() string {
	return d.Legacy.GetSingularName()
}

func (d *DualWriterMode1) NamespaceScoped() bool {
	return d.Legacy.NamespaceScoped()
}

func (d *DualWriterMode1) New() runtime.Object {
	return d.Legacy.New()
}

func (d *DualWriterMode1) NewList() runtime.Object {
	return d.Storage.NewList()
}

func (d *DualWriterMode1) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return d.Legacy.ConvertToTable(ctx, object, tableOptions)
}
