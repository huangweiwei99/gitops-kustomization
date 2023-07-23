/*
Copyright 2022 The open-podcasts Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/SlyMarbo/rss"
	strip "github.com/grokify/html-strip-tags-go"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/opensource-f2f/open-podcasts/api/osf2f.my.domain/v1alpha1"
)

// RSSReconciler reconciles a RSS object
type RSSReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=osf2f.my.domain,resources=rsses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=osf2f.my.domain,resources=rsses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=osf2f.my.domain,resources=rsses/finalizers,verbs=update
//+kubebuilder:rbac:groups=osf2f.my.domain,resources=episodes,verbs=get;create;update;patch;delete
//+kubebuilder:rbac:groups=osf2f.my.domain,resources=categories,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=osf2f.my.domain,resources=authors,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *RSSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	_ = log.FromContext(ctx)

	rssObj := &v1alpha1.RSS{}
	if err = r.Client.Get(ctx, req.NamespacedName, rssObj); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}

	if err = r.linkToCategory(rssObj); err != nil {
		return
	}

	//if err = r.linkToAuthor(rssObj); err != nil {
	//	return
	//}

	address := rssObj.Spec.Address
	if address == "" {
		result = ctrl.Result{RequeueAfter: time.Minute}
		err = r.errorAndRecord(rssObj, v1.EventTypeWarning, "Failed to fetch RSS",
			fmt.Sprintf("the address of the RSS: %s is empty", req.NamespacedName.String()))
		return
	}

	result, err = r.fetchByRSS(address, rssObj)
	return
}

//func (r *RSSReconciler) linkToAuthor(rss *v1alpha1.RSS) (err error) {
//	ctx := context.Background()
//	authorName := rss.Spec.Author
//
//	py := pinyin.New()
//	py.Split = "-"
//	py.Upper = false
//
//	authorNamePy, _ := py.Convert(authorName)
//
//	author := &v1alpha1.Author{
//		Spec: v1alpha1.AuthorSpec{
//			Name: authorName,
//		},
//	}
//	author.SetName(authorNamePy)
//	AddOwnerReference(author, rss.TypeMeta, rss.ObjectMeta)
//
//	if err = r.Get(ctx, types.NamespacedName{Name: authorNamePy}, author); err != nil {
//		if errors.IsNotFound(err) {
//			// create it if is not found
//			err = r.Create(ctx, author)
//		}
//	} else {
//		AddOwnerReference(author, rss.TypeMeta, rss.ObjectMeta)
//		err = r.Update(ctx, author)
//	}
//	return
//}

func (r *RSSReconciler) linkToCategory(rss *v1alpha1.RSS) (err error) {
	ctx := context.Background()
	categories := rss.Spec.Categories
	if len(categories) == 0 {
		return
	}

	for i := range categories {
		categoryName := strings.ToLower(categories[i])
		category := &v1alpha1.Category{}
		category.SetName(categoryName)
		AddOwnerReference(category, rss.TypeMeta, rss.ObjectMeta)

		if err = r.Get(ctx, types.NamespacedName{Name: categoryName}, category); err != nil {
			if errors.IsNotFound(err) {
				// create it if is not found
				err = r.Create(ctx, category)
			} else {
				return
			}
		} else {
			AddOwnerReference(category, rss.TypeMeta, rss.ObjectMeta)
			err = r.Update(ctx, category)
		}
	}
	return
}

func AddOwnerReference(object metav1.Object, typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) {
	SetOwnerReference(object, metav1.OwnerReference{
		Kind:       typeMeta.Kind,
		APIVersion: typeMeta.APIVersion,
		Name:       objectMeta.Name,
		UID:        objectMeta.UID,
	})
}

func SetOwnerReference(object metav1.Object, ownerRef metav1.OwnerReference) {
	allRefs := object.GetOwnerReferences()
	if len(allRefs) == 0 {
		object.SetOwnerReferences([]metav1.OwnerReference{ownerRef})
	} else {
		for i, ref := range allRefs {
			if ref.Name == ownerRef.Name && ref.Kind == ownerRef.Kind &&
				ref.APIVersion == ownerRef.APIVersion {
				allRefs[i] = ownerRef
				return
			}
		}

		object.SetOwnerReferences(append(object.GetOwnerReferences(), ownerRef))
	}
}

func (r *RSSReconciler) fetchByRSS(address string, rssObject *v1alpha1.RSS) (result ctrl.Result, err error) {
	var feed *rss.Feed
	if feed, err = rss.Fetch(address); err != nil {
		err = r.errorAndRecord(rssObject, v1.EventTypeWarning, "Failed to fetch RSS",
			fmt.Sprintf("failed to fetch RSS by address: %s, error is %v", address, err))
		return
	}

	rssObject.Spec.Title = strings.TrimSpace(feed.Title)
	rssObject.Spec.Language = feed.Language
	rssObject.Spec.Author = strings.TrimSpace(feed.Author)
	rssObject.Spec.Description = strings.TrimSpace(strip.StripTags(feed.Description))
	rssObject.Spec.Link = getFixedLink(address, feed)
	rssObject.Spec.Categories = removeDuplicateStr(feed.Categories)
	if feed.Image != nil {
		if feed.Image.URL != "" {
			rssObject.Spec.Image = feed.Image.URL
		} else if feed.Image.Href != "" {
			rssObject.Spec.Image = feed.Image.Href
		}
	}
	if err = r.updateRSS(rssObject.DeepCopy()); err != nil {
		return
	}

	if err = r.storeEpisodes(feed.Items, rssObject.ObjectMeta); err == nil {
		err = r.setLastUpdateTime(rssObject.Namespace, rssObject.Name)
	}

	if err == nil {
		result = ctrl.Result{
			RequeueAfter: feed.Refresh.Sub(time.Now()),
		}
	}
	return
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func getFixedLink(source string, feed *rss.Feed) string {
	if strings.HasPrefix(source, feed.Link) && strings.HasPrefix(source, "http://www.ximalaya.com") {
		return strings.TrimSuffix(source, ".xml")
	}
	return feed.Link
}

func (r *RSSReconciler) updateRSS(newRSS *v1alpha1.RSS) (err error) {
	existingRSS := &v1alpha1.RSS{}
	if err = r.Client.Get(context.Background(), types.NamespacedName{
		Namespace: newRSS.Namespace,
		Name:      newRSS.Name,
	}, existingRSS); err != nil {
		return
	}

	existingRSS.Spec = newRSS.Spec
	err = r.Client.Update(context.Background(), existingRSS)
	return
}

func (r *RSSReconciler) setLastUpdateTime(ns, name string) (err error) {
	rssObj := &v1alpha1.RSS{}
	if err = r.Client.Get(context.Background(), types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, rssObj); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}
	rssObj.Status.LastUpdateTime = metav1.NewTime(time.Now())
	err = r.Client.Status().Update(context.Background(), rssObj)
	return
}

func (r *RSSReconciler) storeEpisodes(items []*rss.Item, meta metav1.ObjectMeta) (err error) {
	for i, _ := range items {
		rssMeta := meta.DeepCopy()
		episodeName := fmt.Sprintf("%s-%d", meta.Name, i)

		var created bool
		if created, err = r.storeEpisode(items[i], rssMeta, episodeName); err != nil {
			return
		}

		if created {
			go func() {
				time.Sleep(time.Second * 3)
				_ = r.recordNewEpisodeEvent(meta.Name, episodeName, meta.Namespace)
			}()
		}
	}
	return
}

func (r *RSSReconciler) recordNewEpisodeEvent(rssName, episodeName, ns string) (err error) {
	rssObj := &v1alpha1.RSS{}
	if err = r.Get(context.Background(), types.NamespacedName{
		Namespace: ns,
		Name:      rssName,
	}, rssObj); err == nil {
		episode := &v1alpha1.Episode{}
		if err = r.Get(context.Background(), types.NamespacedName{
			Namespace: ns,
			Name:      episodeName,
		}, episode); err == nil {
			r.recorder.Eventf(rssObj, v1.EventTypeNormal, episodeName, "《%s》更新节目了: '%s'",
				rssObj.Spec.Title, episode.Spec.Title)
		}
	}
	return
}

func (r *RSSReconciler) storeEpisode(item *rss.Item, meta *metav1.ObjectMeta, episodeName string) (created bool, err error) {
	episode := &v1alpha1.Episode{}
	if err = r.Client.Get(context.Background(), types.NamespacedName{
		Namespace: meta.Namespace,
		Name:      episodeName,
	}, episode); err != nil {
		episode := &v1alpha1.Episode{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: meta.Namespace,
				Name:      episodeName,
				Labels: map[string]string{
					"rss": meta.Name,
				},
				OwnerReferences: []metav1.OwnerReference{{
					Name:       meta.Name,
					UID:        meta.UID,
					Kind:       "RSS",
					APIVersion: "osf2f.my.domain/v1alpha1",
				}},
			},
		}
		created = true
		updateEpisode(episode, item)
		err = r.Client.Create(context.Background(), episode)
	} else {
		episode.OwnerReferences = []metav1.OwnerReference{{
			Name:       meta.Name,
			UID:        meta.UID,
			Kind:       "RSS",
			APIVersion: "osf2f.my.domain/v1alpha1",
		}}
		episode.Labels = map[string]string{
			"rss": meta.Name,
		}
		updateEpisode(episode, item)
		err = r.Client.Update(context.Background(), episode)
	}
	return
}

func updateEpisode(episode *v1alpha1.Episode, item *rss.Item) {
	episode.Spec.Title = item.Title
	episode.Spec.Summary = strings.TrimSpace(strip.StripTags(item.Summary))
	episode.Spec.Content = item.Content
	episode.Spec.Link = item.Link
	episode.Spec.Date = metav1.NewTime(item.Date)

	if len(item.Enclosures) > 0 {
		episode.Spec.AudioSource = item.Enclosures[0].URL
		episode.Spec.AudioType = item.Enclosures[0].Type
		episode.Spec.AudioLength = item.Enclosures[0].Length
	}

	if item.Image != nil {
		if item.Image.URL != "" {
			episode.Spec.CoverImage = item.Image.URL
		} else if item.Image.Href != "" {
			episode.Spec.CoverImage = item.Image.Href
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RSSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("rss")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.RSS{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *RSSReconciler) errorAndRecord(object runtime.Object, eventType, reason, msg string) error {
	r.recorder.Eventf(object, eventType, reason, msg)
	return fmt.Errorf(msg)
}
