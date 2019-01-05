// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timestampname

func cr3ExtractMetadataCreationTimestamp(in reader) string {
	moovIn, err := quicktimeSearchBox(in, "moov")
	log.fatalityCheck(err, "failed to find moov box: %s, %v", in.Name(), err)
	canonBox, err := quicktimeSearchUuidBox(moovIn, "85c0b687820f11e08111f4ce462b6a48")
	// cmt1, err := quicktimeSearchBox(canonBox, "CMT1")
	// log.fatalityCheck(err, "failed to find CMT1 box: %s, %v", in.Name(), err)
	// var cmt1CreationTime = tiffExtractMetadataCreationTimestamp(cmt1, 0)
	// return cmt1CreationTime
	cmt2, err := quicktimeSearchBox(canonBox, "CMT2")
	log.fatalityCheck(err, "failed to find CMT2 box: %s, %v", in.Name(), err)
	var cmt2CreationTime = tiffExtractMetadataCreationTimestamp(cmt2, 0)
	return cmt2CreationTime

	// var err error
	// //var offset int64
	// var atomLength uint32
	// var atomType = make([]byte, 4)
	// log.info("\n")
	// for {
	// 	var nextAtomOffset int64
	// 	var atomLargeLength uint64
	// 	err = binary.Read(in, binary.BigEndian, &atomLength)
	// 	log.fatalityCheck(err, "failed to read atom length: %v", err)
	// 	_, err = io.ReadFull(in, atomType)
	// 	log.fatalityCheck(err, "failed to read atom type: %v")
	// 	// checking for large box:
	// 	if atomLength == 1 {
	// 		err = binary.Read(in, binary.BigEndian, &atomLargeLength)
	// 		nextAtomOffset += int64(atomLargeLength - 4 - 4 - 8) // length, type and large length
	// 		// offset += int64(atomLargeLength)
	// 	} else {
	// 		atomLargeLength = uint64(atomLength)        // for print only
	// 		nextAtomOffset += int64(atomLength - 4 - 4) // length and type 4 bytes each
	// 		// offset += int64(atomLength)
	// 	}
	// 	log.info("found section %s:%d\n", string(atomType), atomLargeLength)
	// 	// checking for last box:
	// 	if atomLength == 0 {
	// 		break
	// 	}
	// 	in.Seek(nextAtomOffset, 1)
	// 	// file.Seek(offset, 0)
	// }
}
