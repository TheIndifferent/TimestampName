// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

func cr3ExtractMetadataCreationTimestamp(in reader) string {
	moovIn, err := quicktimeSearchBox(in, "moov")
	CatchFile(err, in.Name(), "failed to find moov box")
	canonBox, err := quicktimeSearchUuidBox(moovIn, "85c0b687820f11e08111f4ce462b6a48")

	cmt1, err := quicktimeSearchBox(canonBox, "CMT1")
	CatchFile(err, in.Name(), "failed to find CMT1 box")
	var cmt1CreationTime = tiffExtractMetadataCreationTimestamp(cmt1, 0)

	_, err = canonBox.Seek(0, 0)
	CatchFile(err, in.Name(), "failed to rewind")
	cmt2, err := quicktimeSearchBox(canonBox, "CMT2")
	CatchFile(err, in.Name(), "failed to find CMT2 box")
	var cmt2CreationTime = tiffExtractMetadataCreationTimestamp(cmt2, 0)
	if cmt1CreationTime < cmt2CreationTime {
		return cmt1CreationTime
	}
	return cmt2CreationTime
}
