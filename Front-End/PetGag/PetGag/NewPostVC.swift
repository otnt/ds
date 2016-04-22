//
//  NewPostVC.swift
//  PetGag
//
//  Created by Rajdeep Mann on 4/18/16.
//  Copyright © 2016 PetGag. All rights reserved.
//

import UIKit

class NewPostVC: UIViewController, UINavigationControllerDelegate,UIImagePickerControllerDelegate {
    var imagePicker: UIImagePickerController!
    @IBOutlet weak var imageView: UIImageView?
    @IBOutlet weak var textField: UITextField?
    let petGagAPI = PetGagAPI.sharedInstane;

    override func viewDidLoad() {
        super.viewDidLoad()
        imagePicker =  UIImagePickerController()
        imagePicker.delegate = self

        // Do any additional setup after loading the view.
    }

    override func didReceiveMemoryWarning() {
        super.didReceiveMemoryWarning()
        // Dispose of any resources that can be recreated.
    }
    

    /*
    // MARK: - Navigation

    // In a storyboard-based application, you will often want to do a little preparation before navigation
    override func prepareForSegue(segue: UIStoryboardSegue, sender: AnyObject?) {
        // Get the new view controller using segue.destinationViewController.
        // Pass the selected object to the new view controller.
    }
    */

    
    @IBAction func pickImageFromGallery(sender: UIButton) {
        imagePicker.sourceType = .PhotoLibrary
        
        presentViewController(imagePicker, animated: true, completion: nil)

    }
    
    @IBAction func useCamera(sender: UIButton) {
        
        if UIImagePickerController.isSourceTypeAvailable(.Camera) {
            imagePicker.sourceType = .Camera
            
            presentViewController(imagePicker, animated: true, completion: nil)
        }
        

    }
  
    
    @IBAction func postGag(sender: UIButton) {
        
        if self.imageView?.image != nil {
                            
            
            let gagInfo = ["ImageFile":(self.imageView?.image)!] as NSMutableDictionary;
            
            //gagInfo["NEW_TAG"]="NEW_VALUE"
            
            petGagAPI.postGag(gagInfo, completionHandler: { (trendingObjects) in
                

            });
        
        }
        
    }

    
    func imagePickerController(picker: UIImagePickerController, didFinishPickingMediaWithInfo info: [String : AnyObject]){
        imagePicker.dismissViewControllerAnimated(true, completion: nil)
        self.imageView!.image = info[UIImagePickerControllerOriginalImage] as? UIImage
    }
    
     func imagePickerControllerDidCancel(picker: UIImagePickerController){
        imagePicker.dismissViewControllerAnimated(true, completion: nil)
    }
    
    
}
